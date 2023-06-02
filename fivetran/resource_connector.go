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

func resourceConnectorCreate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
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

	createConnectorService.ConfigCustom(resourceConnectorUpdateCustomConfig(resourceData))

	createConnectorService.Auth(resourceConnectorCreateAuth(resourceData.Get("auth").([]interface{})))
	createConnectorService.AuthCustom(resourceConnectorUpdateCustomAuth(resourceData))

	resp, err := createConnectorService.DoCustomMerged(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	resourceData.SetId(resp.Data.ID)
	resourceConnectorRead(ctx, resourceData, m)

	return diags
}

func resourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

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

	// msi stands for Map String Interface
	msi := getConnectorRead(&currentConfig, resp, 1)

	currentService := resourceData.Get("service").(string)

	// Ignore service change for migrated `adwords` connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(msi, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	}

	for k, v := range msi {
		if err := resourceData.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	resourceData.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
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

	modifyConnectorService.ConfigCustom(resourceConnectorUpdateCustomConfig(resourceData))
	modifyConnectorService.Auth(resourceConnectorCreateAuth(resourceData.Get("auth").([]interface{})))
	modifyConnectorService.AuthCustom(resourceConnectorUpdateCustomAuth(resourceData))

	resp, err := modifyConnectorService.DoCustomMerged(ctx)
	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorRead(ctx, resourceData, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := resourceData.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorRead(ctx, resourceData, m)
}

func resourceConnectorDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	deleteConnectorService := client.NewConnectorDelete()

	resp, err := deleteConnectorService.ConnectorID(resourceData.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	resourceData.SetId("")

	return diags
}

func resourceConnectorUpdateCustomConfig(resourceData *schema.ResourceData) *map[string]interface{} {
	configResult := make(map[string]interface{})

	var resourceConfigs = resourceData.Get("config").([]interface{})

	if len(resourceConfigs) < 1 {
		return &configResult
	}
	if resourceConfigs[0] == nil {
		return &configResult
	}

	responseConfig := resourceConfigs[0].(map[string]interface{})

	fields := getFields()

	for fieldName, fieldSchema := range fields {
		if fieldSchema.Type == schema.TypeSet || fieldSchema.Type == schema.TypeList {
			if values := responseConfig[fieldName].([]interface{}); len(values) > 0 {
				if mapValues, ok := values[0].(map[string]interface{}); ok {
					for childPropertyKey, _ := range mapValues {
						if childPropertyValues, ok := mapValues[childPropertyKey].(*schema.Set); ok && len(childPropertyValues.List()) > 0 {
							mapValues[childPropertyKey] = childPropertyValues
							continue
						}
					}
					values[0] = mapValues
					configResult[fieldName] = values
				} else {
					configResult[fieldName] = xInterfaceStrXStr(values)
				}
				continue
			}
			if values, ok := responseConfig[fieldName].(*schema.Set); ok {
				setValues := values.List()

				fmt.Printf("this property is now:%v", setValues)
			}

			if values, ok := responseConfig[fieldName].([]string); ok {
				configResult[fieldName] = xStrXInterface(values)
				continue
			}
		}
		if value, ok := responseConfig[fieldName].(string); ok && value != "" {
			switch fieldSchema.Type {
			case schema.TypeBool:
				configResult[fieldName] = strToBool(value)
			case schema.TypeInt:
				configResult[fieldName] = strToInt(value)
			default:
				configResult[fieldName] = value
			}
			continue
		}
		if value, ok := responseConfig[fieldName].(bool); ok {
			configResult[fieldName] = value
			continue
		}
		if value, ok := responseConfig[fieldName].(int); ok {
			configResult[fieldName] = value
			continue
		}
	}

	return &configResult
}

func resourceConnectorUpdateCustomAuth(d *schema.ResourceData) *map[string]interface{} {
	authMap := make(map[string]interface{})

	var auth = d.Get("auth").([]interface{})

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
