package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:        getConnectorSchema(false, 1),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceConnectorV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceconnectorInstanceStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceConnectorV0() *schema.Resource {
	return &schema.Resource{
		Schema: getConnectorSchema(false, 0),
	}
}

func resourceconnectorInstanceStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	// These fields are managed by `fivetran_connector_schedule` resource
	delete(rawState, "sync_frequency")
	delete(rawState, "schedule_type")
	delete(rawState, "paused")
	delete(rawState, "pause_after_trial")
	delete(rawState, "daily_sync_time")

	// These fields doesn't make sense for resource as they are mutable
	delete(rawState, "status")
	delete(rawState, "succeeded_at")
	delete(rawState, "failed_at")
	delete(rawState, "service_version")

	return rawState, nil
}

func resourceConnectorCreate(ctx context.Context, resourceData *schema.ResourceData, clientInterface interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := clientInterface.(*fivetran.Client)
	createConnectorService := client.NewConnectorCreate()

	createConnectorService.GroupID(resourceData.Get("group_id").(string))

	currentService := resourceData.Get("service").(string)

	if currentService == "adwords" {
		return newDiagAppend(diags, diag.Error, "create error", "service `adwords` has been deprecated, use `google_ads` instead")
	}

	createConnectorService.Service(currentService)

	// new connector always in paused state
	// `fivetran_connector_schedule` should be used for schedule management
	createConnectorService.Paused(true)
	createConnectorService.PauseAfterTrial(true)

	createConnectorService.TrustCertificates(strToBool(resourceData.Get("trust_certificates").(string)))
	createConnectorService.TrustFingerprints(strToBool(resourceData.Get("trust_fingerprints").(string)))
	createConnectorService.RunSetupTests(strToBool(resourceData.Get("run_setup_tests").(string)))

	createConnectorService.ConfigCustom(resourceConnectorUpdateCustomConfig(resourceData, true))

	createConnectorService.Auth(resourceConnectorCreateAuth(resourceData.Get("auth").([]interface{})))
	createConnectorService.AuthCustom(resourceConnectorUpdateCustomAuth(resourceData))

	resp, err := createConnectorService.DoCustom(ctx)
	//Ovde puca
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	resourceData.SetId(resp.Data.ID)
	resourceConnectorRead(ctx, resourceData, clientInterface)

	return diags
}

func resourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, clientInterface interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := clientInterface.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(resourceData.Get("id").(string)).DoCustom(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			resourceData.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	currentConfig := resourceData.Get("config").([]interface{})

	dataBucket := getConnectorRead(&currentConfig, resp, 1)

	currentService := resourceData.Get("service").(string)

	// Ignore service change for migrated `adwords` connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(dataBucket, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	}

	for k, v := range dataBucket {
		if err := resourceData.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	resourceData.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, resourceData *schema.ResourceData, clientInterface interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := clientInterface.(*fivetran.Client)
	modifyConnectorService := client.NewConnectorModify()

	modifyConnectorService.ConnectorID(resourceData.Get("id").(string))

	if resourceData.HasChange("sync_frequency") {
		modifyConnectorService.SyncFrequency(strToInt(resourceData.Get("sync_frequency").(string)))
	}
	if resourceData.HasChange("trust_certificates") {
		modifyConnectorService.TrustCertificates(strToBool(resourceData.Get("trust_certificates").(string)))
	}
	if resourceData.HasChange("trust_fingerprints") {
		modifyConnectorService.TrustFingerprints(strToBool(resourceData.Get("trust_fingerprints").(string)))
	}
	if resourceData.HasChange("run_setup_tests") {
		modifyConnectorService.RunSetupTests(strToBool(resourceData.Get("run_setup_tests").(string)))
	}
	if resourceData.HasChange("paused") {
		modifyConnectorService.Paused(strToBool(resourceData.Get("paused").(string)))
	}
	if resourceData.HasChange("pause_after_trial") {
		modifyConnectorService.PauseAfterTrial(strToBool(resourceData.Get("pause_after_trial").(string)))
	}
	if resourceData.Get("sync_frequency") == "1440" && resourceData.HasChange("daily_sync_time") {
		modifyConnectorService.DailySyncTime(resourceData.Get("daily_sync_time").(string))
	}

	modifyConnectorService.ConfigCustom(resourceConnectorUpdateCustomConfig(resourceData, false))
	modifyConnectorService.Auth(resourceConnectorCreateAuth(resourceData.Get("auth").([]interface{})))
	modifyConnectorService.AuthCustom(resourceConnectorUpdateCustomAuth(resourceData))

	resp, err := modifyConnectorService.DoCustom(ctx)
	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorRead(ctx, resourceData, clientInterface)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := resourceData.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorRead(ctx, resourceData, clientInterface)
}

func resourceConnectorDelete(ctx context.Context, resourceData *schema.ResourceData, clientInterface interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := clientInterface.(*fivetran.Client)
	deleteConnectorService := client.NewConnectorDelete()

	resp, err := deleteConnectorService.ConnectorID(resourceData.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	resourceData.SetId("")

	return diags
}

func resourceConnectorUpdateCustomConfig(resourceData *schema.ResourceData, create bool) *map[string]interface{} {
	var resourceConfigs = resourceData.Get("config").([]interface{})

	if len(resourceConfigs) < 1 {
		return &map[string]interface{}{}
	}
	if resourceConfigs[0] == nil {
		return &map[string]interface{}{}
	}

	responseConfig := resourceConfigs[0].(map[string]interface{})

	fields := getFields()

	config := createConfig(responseConfig, fields)

	if create {
		var destinationSchema = resourceData.Get("destination_schema").([]interface{})
		schema := destinationSchema[0].(map[string]interface{})
		if v := schema["name"].(string); v != "" {
			config["schema"] = v
		}
		if v := schema["table"].(string); v != "" {
			config["table"] = v
		}
		if v := schema["prefix"].(string); v != "" {
			config["schema_prefix"] = v
		}
	}

	return &config
}

func createConfig(responseConfig map[string]interface{}, fields map[string]*schema.Schema) map[string]interface{} {
	config := make(map[string]interface{})

	for fieldName, fieldSchema := range fields {
		if _, ok := responseConfig[fieldName]; !ok {
			continue
		}

		if fieldSchema.Type == schema.TypeSet || fieldSchema.Type == schema.TypeList {
			if values := responseConfig[fieldName].(*schema.Set).List(); len(values) > 0 {
				if mapValues, ok := values[0].(map[string]interface{}); ok {
					for childPropertyKey, _ := range mapValues {
						if childPropertyValues, ok := mapValues[childPropertyKey].(*schema.Set); ok {
							mapValues[childPropertyKey] = childPropertyValues.List()
							continue
						}
					}
					values[0] = mapValues
					config[fieldName] = values
				} else {
					config[fieldName] = xInterfaceStrXStr(values)
				}
				continue
			}

			if values, ok := responseConfig[fieldName].([]string); ok {
				config[fieldName] = xStrXInterface(values)
				continue
			}
		}

		if value, ok := responseConfig[fieldName].(string); ok && value != "" {
			switch fieldSchema.Type {
			case schema.TypeBool:
				config[fieldName] = strToBool(value)
			case schema.TypeInt:
				config[fieldName] = strToInt(value)
			default:
				config[fieldName] = value
			}
			continue
		}
		if value, ok := responseConfig[fieldName].(bool); ok {
			config[fieldName] = value
			continue
		}
		if value, ok := responseConfig[fieldName].(int); ok {
			config[fieldName] = value
			continue
		}
	}
	return config
}

func resourceConnectorUpdateCustomAuth(resourceData *schema.ResourceData) *map[string]interface{} {
	authMap := make(map[string]interface{})

	var auth = resourceData.Get("auth").([]interface{})

	if len(auth) < 1 {
		return &authMap
	}
	if auth[0] == nil {
		return &authMap
	}

	return &authMap
}

func resourceConnectorCreateAuth(auth []interface{}) *fivetran.ConnectorAuth {
	fivetranAuth := fivetran.NewConnectorAuth()

	if len(auth) < 1 {
		return fivetranAuth
	}
	if auth[0] == nil {
		return fivetranAuth
	}

	a := auth[0].(map[string]interface{})

	if v := a["client_access"].([]interface{}); len(v) > 0 {
		fivetranAuth.ClientAccess(resourceConnectorCreateAuthClientAccess(v))
	}
	if v := a["refresh_token"].(string); v != "" {
		fivetranAuth.RefreshToken(v)
	}
	if v := a["access_token"].(string); v != "" {
		fivetranAuth.AccessToken(v)
	}
	if v := a["realm_id"].(string); v != "" {
		fivetranAuth.RealmID(v)
	}

	return fivetranAuth
}

func resourceConnectorCreateAuthClientAccess(clientAccess []interface{}) *fivetran.ConnectorAuthClientAccess {
	fivetranAuthClientAccess := fivetran.NewConnectorAuthClientAccess()

	if len(clientAccess) < 1 {
		return fivetranAuthClientAccess
	}
	if clientAccess[0] == nil {
		return fivetranAuthClientAccess
	}

	ca := clientAccess[0].(map[string]interface{})
	if v := ca["client_id"].(string); v != "" {
		fivetranAuthClientAccess.ClientID(v)
	}
	if v := ca["client_secret"].(string); v != "" {
		fivetranAuthClientAccess.ClientSecret(v)
	}
	if v := ca["user_agent"].(string); v != "" {
		fivetranAuthClientAccess.UserAgent(v)
	}
	if v := ca["developer_token"].(string); v != "" {
		fivetranAuthClientAccess.DeveloperToken(v)
	}

	return fivetranAuthClientAccess
}
