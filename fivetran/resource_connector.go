package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnectorLegacy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorLegacyCreate,
		ReadContext:   resourceConnectorLegacyRead,
		UpdateContext: resourceConnectorLegacyUpdate,
		DeleteContext: resourceConnectorLegacyDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:        connectorSchemaLegacy(false, 2),
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceConnectorLegacyV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceconnectorInstanceStateUpgradeV0Legacy,
				Version: 0,
			},
			{
				Type:    resourceConnectorLegacyV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceconnectorInstanceStateUpgradeV1Legacy,
				Version: 1,
			},
		},
	}
}

func resourceConnectorLegacyV0() *schema.Resource {
	return &schema.Resource{
		Schema: connectorSchemaLegacy(false, 0),
	}
}

func resourceConnectorLegacyV1() *schema.Resource {
	return &schema.Resource{
		Schema: connectorSchemaLegacy(false, 1),
	}
}

func resourceconnectorInstanceStateUpgradeV1Legacy(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if c, ok := rawState["config"].([]interface{}); ok && len(c) == 1 {
		// The field `servers` had wrong type and couldn't be used effectively
		// Now we should just override it in state object to avoid migration collision
		if config, ok := c[0].(map[string]interface{}); ok {
			config["servers"] = nil
		}
	}

	return rawState, nil
}

func resourceconnectorInstanceStateUpgradeV0Legacy(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
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

func resourceConnectorLegacyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorCreate()

	svc.GroupID(d.Get("group_id").(string))

	currentService := d.Get("service").(string)

	if currentService == "adwords" {
		return newDiagAppend(diags, diag.Error, "create error", "service `adwords` has been deprecated, use `google_ads` instead")
	}

	svc.Service(currentService)

	// new connector always in paused state
	// `fivetran_connector_schedule` should be used for schedule management
	svc.Paused(true)
	svc.PauseAfterTrial(true)

	svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))

	destination_schema := d.Get("destination_schema").([]interface{})[0].(map[string]interface{})

	config := resourceConnectorLegacyUpdateCustomConfig(d)

	if v := destination_schema["name"].(string); v != "" {
		config["schema"] = v
	}
	if v := destination_schema["table"].(string); v != "" {
		config["table"] = v
	}
	if v := destination_schema["prefix"].(string); v != "" {
		config["schema_prefix"] = v
	}

	svc.ConfigCustom(&config)
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustom(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorLegacyRead(ctx, d, m)

	return diags
}

func resourceConnectorLegacyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	currentConfig := d.Get("config").([]interface{})

	msi := connectorRead(&currentConfig, resp, 1)

	currentService := d.Get("service").(string)

	// Ignore service change for migrated `adwords` connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(msi, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	}

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorLegacyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorModify()

	svc.ConnectorID(d.Get("id").(string))

	if d.HasChange("sync_frequency") {
		svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	}
	if d.HasChange("trust_certificates") {
		svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	}
	if d.HasChange("trust_fingerprints") {
		svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	}
	if d.HasChange("run_setup_tests") {
		svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))
	}
	if d.HasChange("paused") {
		svc.Paused(strToBool(d.Get("paused").(string)))
	}
	if d.HasChange("pause_after_trial") {
		svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	}
	if d.Get("sync_frequency") == "1440" && d.HasChange("daily_sync_time") {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	config := resourceConnectorLegacyUpdateCustomConfig(d)

	svc.ConfigCustom(&config)
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustom(ctx)
	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorLegacyRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorLegacyRead(ctx, d, m)
}

func resourceConnectorLegacyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorDelete()

	resp, err := svc.ConnectorID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

func resourceConnectorLegacyUpdateCustomConfig(d *schema.ResourceData) map[string]interface{} {
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
