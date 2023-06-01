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

	svc.ConfigCustom(resourceConnectorUpdateCustomConfig(d))

	svc.Auth(resourceConnectorCreateAuth(d.Get("auth").([]interface{})))
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustomMerged(ctx)
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

	//svc.Config(resourceConnectorUpdateConfig(d))
	svc.ConfigCustom(resourceConnectorUpdateCustomConfig(d))
	svc.Auth(resourceConnectorCreateAuth(d.Get("auth").([]interface{})))
	svc.AuthCustom(resourceConnectorUpdateCustomAuth(d))

	resp, err := svc.DoCustomMerged(ctx)
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

func resourceConnectorUpdateCustomConfig(d *schema.ResourceData) *map[string]interface{} {
	configResult := make(map[string]interface{})

	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return &configResult
	}
	if config[0] == nil {
		return &configResult
	}

	responseConfig := config[0].(map[string]interface{})

	//services := getAvailableServiceIds()

	fields := getFields()

	for fieldName, fieldSchema := range fields {
		if fieldSchema.Type == schema.TypeSet || fieldSchema.Type == schema.TypeList {
			if values := responseConfig[fieldName].([]interface{}); len(values) > 0 {
				if mapValues, ok := values[0].(map[string]interface{}); ok {
					for childPropertyKey, _ := range mapValues {
						if _, ok := mapValues[childPropertyKey].(string); ok {
							continue
						}
						if _, ok := mapValues[childPropertyKey].(bool); ok {
							continue
						}
						if _, ok := mapValues[childPropertyKey].([]interface{}); ok {
							continue
						}
						if childPropertyValues := mapValues[childPropertyKey].(*schema.Set).List(); len(childPropertyValues) > 0 {
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

func resourceConnectorCreateConfig(fivetranConfig *fivetran.ConnectorConfig, destination_schema []interface{}) *fivetran.ConnectorConfig {
	d := destination_schema[0].(map[string]interface{})
	if v := d["name"].(string); v != "" {
		fivetranConfig.Schema(v)
	}
	if v := d["table"].(string); v != "" {
		fivetranConfig.Table(v)
	}
	if v := d["prefix"].(string); v != "" {
		fivetranConfig.SchemaPrefix(v)
	}

	return fivetranConfig
}

func resourceConnectorCreateFunctionSecrets(xi []interface{}) []*fivetran.FunctionSecret {
	functionSecrets := make([]*fivetran.FunctionSecret, len(xi))
	for i, v := range xi {
		vmap := v.(map[string]interface{})
		// As the fields are marked as required in schema we can skip any checks here
		functionSecrets[i] =
			fivetran.NewFunctionSecret().
				Key(vmap["key"].(string)).
				Value(vmap["value"].(string))
	}
	return functionSecrets
}

func resourceConnectorCreateConfigProjectCredentials(xi []interface{}) []*fivetran.ConnectorConfigProjectCredentials {
	projectCredentials := make([]*fivetran.ConnectorConfigProjectCredentials, len(xi))
	for i, v := range xi {
		pc := fivetran.NewConnectorConfigProjectCredentials()
		if project, ok := v.(map[string]interface{})["project"].(string); ok && project != "" {
			pc.Project(project)
		}
		if apiKey, ok := v.(map[string]interface{})["api_key"].(string); ok && apiKey != "" {
			pc.APIKey(apiKey)
		}
		if secretKey, ok := v.(map[string]interface{})["secret_key"].(string); ok && secretKey != "" {
			pc.SecretKey(secretKey)
		}
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func resourceConnectorCreateConfigCustomTables(xi []interface{}) []*fivetran.ConnectorConfigCustomTables {
	customTables := make([]*fivetran.ConnectorConfigCustomTables, len(xi))
	for i, v := range xi {
		ct := fivetran.NewConnectorConfigCustomTables()
		if tableName, ok := v.(map[string]interface{})["table_name"].(string); ok && tableName != "" {
			ct.TableName(tableName)
		}
		if configType, ok := v.(map[string]interface{})["config_type"].(string); ok && configType != "" {
			ct.ConfigType(configType)
		}
		if fields, ok := v.(map[string]interface{})["fields"]; ok {
			ct.Fields(xInterfaceStrXStr(fields.(*schema.Set).List()))
		}
		if breakdowns, ok := v.(map[string]interface{})["breakdowns"]; ok {
			ct.Breakdowns(xInterfaceStrXStr(breakdowns.(*schema.Set).List()))
		}
		if actionBreakdowns, ok := v.(map[string]interface{})["action_breakdowns"]; ok {
			ct.ActionBreakdowns(xInterfaceStrXStr(actionBreakdowns.(*schema.Set).List()))
		}
		if aggregation, ok := v.(map[string]interface{})["aggregation"].(string); ok && aggregation != "" {
			ct.Aggregation(aggregation)
		}
		if actionReportTime, ok := v.(map[string]interface{})["action_report_time"].(string); ok && actionReportTime != "" {
			ct.ActionReportTime(actionReportTime)
		}
		if clickAttributionWindow, ok := v.(map[string]interface{})["click_attribution_window"].(string); ok && clickAttributionWindow != "" {
			ct.ClickAttributionWindow(clickAttributionWindow)
		}
		if viewAttributionWindow, ok := v.(map[string]interface{})["view_attribution_window"].(string); ok && viewAttributionWindow != "" {
			ct.ViewAttributionWindow(viewAttributionWindow)
		}
		if prebuiltReportName, ok := v.(map[string]interface{})["prebuilt_report_name"].(string); ok && prebuiltReportName != "" {
			ct.PrebuiltReportName(prebuiltReportName)
		}
		customTables[i] = ct
	}

	return customTables
}

func resourceConnectorCreateConfigAdobeAnalyticsConfigurations(xi []interface{}) []*fivetran.ConnectorConfigAdobeAnalyticsConfiguration {
	configurations := make([]*fivetran.ConnectorConfigAdobeAnalyticsConfiguration, len(xi))
	for i, v := range xi {
		c := fivetran.NewConnectorConfigAdobeAnalyticsConfiguration()

		if syncMode, ok := v.(map[string]interface{})["sync_mode"].(string); ok && syncMode != "" {
			c.SyncMode(syncMode)
		}
		if metrics, ok := v.(map[string]interface{})["metrics"]; ok {
			c.Metrics(xInterfaceStrXStr(metrics.(*schema.Set).List()))
		}
		if reportSuites, ok := v.(map[string]interface{})["report_suites"]; ok {
			c.ReportSuites(xInterfaceStrXStr(reportSuites.(*schema.Set).List()))
		}
		if segments, ok := v.(map[string]interface{})["segments"]; ok {
			c.Segments(xInterfaceStrXStr(segments.(*schema.Set).List()))
		}
		if elements, ok := v.(map[string]interface{})["elements"]; ok {
			c.Elements(xInterfaceStrXStr(elements.(*schema.Set).List()))
		}
		if calculatedMetrics, ok := v.(map[string]interface{})["calculated_metrics"]; ok {
			c.CalculatedMetrics(xInterfaceStrXStr(calculatedMetrics.(*schema.Set).List()))
		}

		configurations[i] = c
	}

	return configurations
}

func resourceConnectorCreateConfigReports(xi []interface{}) []*fivetran.ConnectorConfigReports {
	reports := make([]*fivetran.ConnectorConfigReports, len(xi))
	for i, v := range xi {
		r := fivetran.NewConnectorConfigReports()
		if table, ok := v.(map[string]interface{})["table"].(string); ok && table != "" {
			r.Table(table)
		}
		if configType, ok := v.(map[string]interface{})["config_type"].(string); ok && configType != "" {
			r.ConfigType(configType)
		}
		if prebuiltReport, ok := v.(map[string]interface{})["prebuilt_report"].(string); ok && prebuiltReport != "" {
			r.PrebuiltReport(prebuiltReport)
		}
		if reportType, ok := v.(map[string]interface{})["report_type"].(string); ok && reportType != "" {
			r.ReportType(reportType)
		}
		if fields, ok := v.(map[string]interface{})["fields"]; ok {
			r.Fields(xInterfaceStrXStr(fields.(*schema.Set).List()))
		}
		if dimensions, ok := v.(map[string]interface{})["dimensions"]; ok {
			r.Dimensions(xInterfaceStrXStr(dimensions.(*schema.Set).List()))
		}
		if metrics, ok := v.(map[string]interface{})["metrics"]; ok {
			r.Metrics(xInterfaceStrXStr(metrics.(*schema.Set).List()))
		}
		if segments, ok := v.(map[string]interface{})["segments"]; ok {
			r.Segments(xInterfaceStrXStr(segments.(*schema.Set).List()))
		}
		if filter, ok := v.(map[string]interface{})["filter"].(string); ok && filter != "" {
			r.Filter(filter)
		}
		reports[i] = r
	}

	return reports
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
