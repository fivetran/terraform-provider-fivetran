package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorCreate,
		ReadContext:          resourceConnectorRead,
		UpdateWithoutTimeout: resourceConnectorUpdate,
		DeleteContext:        resourceConnectorDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:               getConnectorSchema(false, 2),
		SchemaVersion:        2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceConnectorLegacyV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceconnectorInstanceStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceConnectorLegacyV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceconnectorInstanceStateUpgradeV1,
				Version: 1,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceConnectorLegacyV0() *schema.Resource {
	return &schema.Resource{
		Schema: getConnectorSchema(false, 0),
	}
}

func resourceConnectorLegacyV1() *schema.Resource {
	return &schema.Resource{
		Schema: getConnectorSchema(false, 1),
	}
}

func resourceconnectorInstanceStateUpgradeV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if c, ok := rawState["config"].([]interface{}); ok && len(c) == 1 {
		// The field `servers` had wrong type and couldn't be used effectively
		// Now we should just override it in state object to avoid migration collision
		if config, ok := c[0].(map[string]interface{}); ok {
			config["servers"] = nil
		}
	}

	return rawState, nil
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

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorCreate()

	svc.GroupID(d.Get("group_id").(string))

	currentService := d.Get("service").(string)

	if currentService == "adwords" {
		return helpers.NewDiagAppend(diags, diag.Error, "create error", "service `adwords` has been deprecated, use `google_ads` instead")
	}

	svc.Service(currentService)

	// new connector always in paused state
	// `fivetran_connector_schedule` should be used for schedule management
	svc.Paused(true)
	svc.PauseAfterTrial(true)

	svc.TrustCertificates(helpers.StrToBool(d.Get("trust_certificates").(string)))
	svc.TrustFingerprints(helpers.StrToBool(d.Get("trust_fingerprints").(string)))
	svc.RunSetupTests(helpers.StrToBool(d.Get("run_setup_tests").(string)))

	destination_schema := d.Get("destination_schema").([]interface{})[0].(map[string]interface{})

	config := resourceConnectorUpdateCustomConfig(d)

	appendDestinationSchemaFields(destination_schema, config, currentService)

	svc.ConfigCustom(&config)
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := svc.DoCustom(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorRead(ctx, d, m)

	return diags
}

func appendDestinationSchemaField(tfName, apiName string, destinationSchema, config map[string]interface{}, dsf map[string]bool, service string) error {
	_, fieldRequired := dsf[apiName]
	fieldValue := destinationSchema[tfName].(string)
	fieldDefined := (fieldValue != "")
	if fieldRequired {
		if fieldDefined {
			config[apiName] = fieldValue
		} else {
			return fmt.Errorf("field `destination_schema.%v` should be defined for service: %v", tfName, service)
		}
	} else {
		if fieldDefined {
			return fmt.Errorf("field `destination_schema.%v` should not be defined for service: %v", tfName, service)
		}
	}
	return nil
}

func appendDestinationSchemaFields(destinationSchema, config map[string]interface{}, service string) error {
	if dsf, ok := destinationSchemaFields[service]; ok {
		err := appendDestinationSchemaField("name", "schema", destinationSchema, config, dsf, service)
		if err != nil {
			return err
		}
		err = appendDestinationSchemaField("table", "table", destinationSchema, config, dsf, service)
		if err != nil {
			return err
		}
		err = appendDestinationSchemaField("prefix", "schema_prefix", destinationSchema, config, dsf, service)
		if err != nil {
			return err
		}
	} else {
		if v := destinationSchema["name"].(string); v != "" {
			config["schema"] = v
		}
		if v := destinationSchema["table"].(string); v != "" {
			config["table"] = v
		}
		if v := destinationSchema["prefix"].(string); v != "" {
			config["schema_prefix"] = v
		}
	}
	return nil
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).DoCustom(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return helpers.NewDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	currentConfig := d.Get("config").([]interface{})

	msi, err := connectorRead(&currentConfig, resp, 1)

	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "read error", err.Error())
	}

	currentService := d.Get("service").(string)

	// Ignore service change for migrated `adwords` connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		helpers.MapAddStr(msi, "service", "adwords")
		diags = helpers.NewDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	}

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorModify()

	svc.ConnectorID(d.Get("id").(string))

	if d.HasChange("sync_frequency") {
		svc.SyncFrequency(helpers.StrToInt(d.Get("sync_frequency").(string)))
	}
	if d.HasChange("trust_certificates") {
		svc.TrustCertificates(helpers.StrToBool(d.Get("trust_certificates").(string)))
	}
	if d.HasChange("trust_fingerprints") {
		svc.TrustFingerprints(helpers.StrToBool(d.Get("trust_fingerprints").(string)))
	}
	if d.HasChange("run_setup_tests") {
		svc.RunSetupTests(helpers.StrToBool(d.Get("run_setup_tests").(string)))
	}
	if d.HasChange("paused") {
		svc.Paused(helpers.StrToBool(d.Get("paused").(string)))
	}
	if d.HasChange("pause_after_trial") {
		svc.PauseAfterTrial(helpers.StrToBool(d.Get("pause_after_trial").(string)))
	}
	if d.Get("sync_frequency") == "1440" && d.HasChange("daily_sync_time") {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	config := resourceConnectorUpdateCustomConfig(d)

	svc.ConfigCustom(&config)
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	resp, err := svc.DoCustom(ctx)
	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorRead(ctx, d, m)
		return helpers.NewDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorRead(ctx, d, m)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorDelete()

	resp, err := svc.ConnectorID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

func resourceConnectorUpdateCustomConfig(d *schema.ResourceData) map[string]interface{} {
	configMap := make(map[string]interface{})

	var config = d.Get("config").([]interface{})
	var service = d.Get("service").(string)

	if len(config) < 1 || config[0] == nil {
		return configMap
	}

	c := config[0].(map[string]interface{})

	return connectorUpdateCustomConfig(c, service)
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

	a := auth[0].(map[string]interface{})

	if v := a["client_access"].([]interface{}); len(v) > 0 {
		caMap := make(map[string]interface{})
		ca := v[0].(map[string]interface{})
		if cv := ca["client_id"].(string); cv != "" {
			caMap["client_id"] = cv
		}
		if cv := ca["client_secret"].(string); cv != "" {
			caMap["client_secret"] = cv
		}
		if cv := ca["user_agent"].(string); cv != "" {
			caMap["user_agent"] = cv
		}
		if cv := ca["developer_token"].(string); cv != "" {
			caMap["developer_token"] = cv
		}
		authMap["client_access"] = caMap
	}

	if v := a["refresh_token"].(string); v != "" {
		authMap["refresh_token"] = v
	}

	if v := a["access_token"].(string); v != "" {
		authMap["access_token"] = v
	}

	if v := a["realm_id"].(string); v != "" {
		authMap["realm_id"] = v
	}

	return &authMap
}
