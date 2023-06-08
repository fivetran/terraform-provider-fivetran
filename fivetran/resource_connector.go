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
		Schema:        connectorSchema(false, 1),
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
		Schema: connectorSchema(false, 0),
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

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	svc.ConfigCustom(resourceConnectorUpdateCustomConfig(d, true))
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustom(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorRead(ctx, d, m)

	return diags
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

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	svc.ConfigCustom(resourceConnectorUpdateCustomConfig(d, false))
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustom(ctx)

	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorRead(ctx, d, m)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceConnectorUpdateCustomConfig(d *schema.ResourceData, create bool) *map[string]interface{} {
	configMap := make(map[string]interface{})

	if create {
		var destinationSchema = d.Get("destination_schema").([]interface{})

		schema := destinationSchema[0].(map[string]interface{})
		if v := schema["name"].(string); v != "" {
			configMap["schema"] = v
		}
		if v := schema["table"].(string); v != "" {
			configMap["table"] = v
		}
		if v := schema["prefix"].(string); v != "" {
			configMap["schema_prefix"] = v
		}
	}

	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return &configMap
	}
	if config[0] == nil {
		return &configMap
	}

	c := config[0].(map[string]interface{})

	for k, f := range simpleFields {
		if f.fieldValueType == String {
			if v, ok := c[k].(string); ok && v != "" {
				configMap[k] = v
			}
		}
		if f.fieldValueType == StringList {
			if v := c[k].(*schema.Set).List(); len(v) > 0 {
				configMap[k] = xInterfaceStrXStr(v)
			}
		}
		if f.fieldValueType == Integer || f.fieldValueType == Boolean {
			configMap[k] = c[k]
		}
	}

	if v := c["project_credentials"].(*schema.Set).List(); len(v) > 0 {
		configMap["project_credentials"] = v
	}
	if v := c["secrets_list"].(*schema.Set).List(); len(v) > 0 {
		configMap["secrets_list"] = v
	}
	if v := c["custom_tables"].(*schema.Set).List(); len(v) > 0 {
		configMap["custom_tables"] = resourceConnectorCreateConfigCustomTables(v)
	}
	if v := c["reports"].(*schema.Set).List(); len(v) > 0 {
		configMap["reports"] = resourceConnectorCreateConfigReports(v)
	}
	if v := c["adobe_analytics_configurations"].(*schema.Set).List(); len(v) > 0 {
		configMap["adobe_analytics_configurations"] = resourceConnectorCreateConfigAdobeAnalyticsConfigurations(v)
	}

	return &configMap
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
	a := auth[0].(map[string]interface{})

	if v := a["client_access"].([]interface{}); len(v) > 0 {
		authMap["client_access"] = v[0]
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

func resourceConnectorCreateConfigCustomTables(xi []interface{}) []map[string]interface{} {
	customTables := make([]map[string]interface{}, len(xi))
	for i, v := range xi {
		ct := make(map[string]interface{})
		if x, ok := v.(map[string]interface{})["table_name"].(string); ok && x != "" {
			ct["table_name"] = x
		}
		if x, ok := v.(map[string]interface{})["config_type"].(string); ok && x != "" {
			ct["config_type"] = x
		}
		if x, ok := v.(map[string]interface{})["fields"]; ok {
			ct["fields"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["breakdowns"]; ok {
			ct["breakdowns"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["action_breakdowns"]; ok {
			ct["action_breakdowns"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["aggregation"].(string); ok && x != "" {
			ct["aggregation"] = x
		}
		if x, ok := v.(map[string]interface{})["action_report_time"].(string); ok && x != "" {
			ct["action_report_time"] = x
		}
		if x, ok := v.(map[string]interface{})["click_attribution_window"].(string); ok && x != "" {
			ct["click_attribution_window"] = x
		}
		if x, ok := v.(map[string]interface{})["view_attribution_window"].(string); ok && x != "" {
			ct["view_attribution_window"] = x
		}
		if x, ok := v.(map[string]interface{})["prebuilt_report_name"].(string); ok && x != "" {
			ct["prebuilt_report_name"] = x
		}
		customTables[i] = ct
	}

	return customTables
}

func resourceConnectorCreateConfigAdobeAnalyticsConfigurations(xi []interface{}) []map[string]interface{} {
	configurations := make([]map[string]interface{}, len(xi))

	for i, v := range xi {
		c := make(map[string]interface{})

		if x, ok := v.(map[string]interface{})["sync_mode"].(string); ok && x != "" {
			c["sync_mode"] = x
		}
		if x, ok := v.(map[string]interface{})["metrics"]; ok {
			c["metrics"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["report_suites"]; ok {
			c["report_suites"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["segments"]; ok {
			c["segments"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["elements"]; ok {
			c["elements"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["calculated_metrics"]; ok {
			c["calculated_metrics"] = x.(*schema.Set).List()
		}

		configurations[i] = c
	}

	return configurations
}

func resourceConnectorCreateConfigReports(xi []interface{}) []map[string]interface{} {
	reports := make([]map[string]interface{}, len(xi))
	for i, v := range xi {
		r := make(map[string]interface{})
		if x, ok := v.(map[string]interface{})["table"].(string); ok && x != "" {
			r["table"] = x
		}
		if x, ok := v.(map[string]interface{})["config_type"].(string); ok && x != "" {
			r["config_type"] = x
		}
		if x, ok := v.(map[string]interface{})["prebuilt_report"].(string); ok && x != "" {
			r["prebuilt_report"] = x
		}
		if x, ok := v.(map[string]interface{})["report_type"].(string); ok && x != "" {
			r["report_type"] = x
		}
		if x, ok := v.(map[string]interface{})["filter"].(string); ok && x != "" {
			r["filter"] = x
		}
		if x, ok := v.(map[string]interface{})["fields"]; ok {
			r["fields"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["dimensions"]; ok {
			r["dimensions"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["metrics"]; ok {
			r["metrics"] = x.(*schema.Set).List()
		}
		if x, ok := v.(map[string]interface{})["segments"]; ok {
			r["segments"] = x.(*schema.Set).List()
		}
		reports[i] = r
	}

	return reports
}
