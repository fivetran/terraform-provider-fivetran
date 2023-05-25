package fivetran

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnectorAutomatic() *schema.Resource {

	var result = &schema.Resource{
		CreateContext: resourceConnectorAutomaticCreate,
		ReadContext:   resourceConnectorAutomaticRead,
		UpdateContext: resourceConnectorAutomaticUpdate,
		DeleteContext: resourceConnectorAutomaticDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":                 {Type: schema.TypeString, Computed: true},
			"group_id":           {Type: schema.TypeString, Required: true, ForceNew: true},
			"service":            {Type: schema.TypeString, Required: true, ForceNew: true},
			"service_version":    {Type: schema.TypeString, Computed: true},
			"destination_schema": resourceConnectorAutomaticDestinationSchemaSchema(),
			"name":               {Type: schema.TypeString, Computed: true},
			"connected_by":       {Type: schema.TypeString, Computed: true},
			"created_at":         {Type: schema.TypeString, Computed: true},
			"succeeded_at":       {Type: schema.TypeString, Computed: true},
			"failed_at":          {Type: schema.TypeString, Computed: true},
			"sync_frequency":     {Type: schema.TypeString, Required: true},
			"daily_sync_time":    {Type: schema.TypeString, Optional: true},
			"schedule_type":      {Type: schema.TypeString, Computed: true},
			"trust_certificates": {Type: schema.TypeString, Optional: true},
			"trust_fingerprints": {Type: schema.TypeString, Optional: true},
			"run_setup_tests":    {Type: schema.TypeString, Optional: true},
			"paused":             {Type: schema.TypeString, Required: true},
			"pause_after_trial":  {Type: schema.TypeString, Required: true},
			"status":             resourceConnectorAutomaticSchemaStatus(),
			"config":             resourceConnectorAutomaticConfigCreate(),
			"auth":               resourceConnectorAutomaticSchemaAuth(),
			"last_updated":       {Type: schema.TypeString, Computed: true}, // internal
		},
	}
	return result
}

func resourceConnectorAutomaticConfigCreate() *schema.Schema {
	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getCSchemaAndProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			properties[key] = value
		}
	}

	return &schema.Schema{Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: properties,
		},
	}
}

func resourceConnectorAutomaticDestinationSchemaSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Optional: true, ForceNew: true},
				"table":  {Type: schema.TypeString, Optional: true, ForceNew: true},
				"prefix": {Type: schema.TypeString, Optional: true, ForceNew: true},
			},
		},
	}
}

func resourceConnectorAutomaticSchemaStatus() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"setup_state":        {Type: schema.TypeString, Computed: true},
				"sync_state":         {Type: schema.TypeString, Computed: true},
				"update_state":       {Type: schema.TypeString, Computed: true},
				"is_historical_sync": {Type: schema.TypeString, Computed: true},
				"tasks": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
				"warnings": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
			},
		},
	}
}

func resourceConnectorAutomaticSchemaAuth() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Optional: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_access": {Type: schema.TypeList, Optional: true, MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id":       {Type: schema.TypeString, Optional: true},
							"client_secret":   {Type: schema.TypeString, Optional: true, Sensitive: true},
							"user_agent":      {Type: schema.TypeString, Optional: true},
							"developer_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
						},
					},
				},
				"refresh_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
				"access_token":  {Type: schema.TypeString, Optional: true, Sensitive: true},
				"realm_id":      {Type: schema.TypeString, Optional: true, Sensitive: true},
			},
		},
	}
}

func resourceConnectorAutomaticCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorCreate()

	svc.GroupID(d.Get("group_id").(string))

	currentService := d.Get("service").(string)

	if currentService == "adwords" {
		return newDiagAppend(diags, diag.Error, "create error", "service `adwords` has been deprecated, use `google_ads` instead")
	}

	svc.Service(currentService)
	svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))
	svc.Paused(strToBool(d.Get("paused").(string)))
	svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	if d.Get("sync_frequency") == "1440" && d.Get("daily_sync_time").(string) != "" {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	// fivetranConfig := resourceConnectorAutomaticUpdateConfig(d)

	// svc.Config(resourceConnectorAutomaticCreateConfig(fivetranConfig, d.Get("destination_schema").([]interface{})))
	svc.ConfigCustom(resourceConnectorAutomaticUpdateCustomConfig(d))

	svc.Auth(resourceConnectorAutomaticCreateAuth(d.Get("auth").([]interface{})))
	svc.AuthCustom(resourceConnectorAutomaticUpdateCustomAuth(d))

	resp, err := svc.DoCustomMerged(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorAutomaticRead(ctx, d, m)

	return diags
}

func resourceConnectorAutomaticRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).DoCustomMerged(ctx)
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
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)

	currentService := d.Get("service").(string)

	// ignore service change for migrated `adwords connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(msi, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	} else {
		mapAddStr(msi, "service", resp.Data.Service)
	}

	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", readDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	if *resp.Data.SyncFrequency == 1440 {
		mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	} else {
		mapAddStr(msi, "daily_sync_time", d.Get("daily_sync_time").(string))
	}
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", resourceConnectorReadStatus(&resp))
	currentConfig := d.Get("config").([]interface{})
	upstreamConfig := resourceConnectorAutomaticReadConfig(&resp, currentConfig)

	if len(upstreamConfig) > 0 {
		mapAddXInterface(msi, "config", upstreamConfig)
	}

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorAutomaticUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	//svc.Config(resourceConnectorAutomaticUpdateConfig(d))
	svc.ConfigCustom(resourceConnectorAutomaticUpdateCustomConfig(d))
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

func resourceConnectorAutomaticDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceConnectorAutomaticUpdateCustomConfig(d *schema.ResourceData) *map[string]interface{} {
	configResult := make(map[string]interface{})

	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return &configResult
	}
	if config[0] == nil {
		return &configResult
	}

	responseConfig := config[0].(map[string]interface{})

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getCSchemaAndProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			properties[key] = value
		}
	}

	for property, propertySchema := range properties {
		if property == "api_type" {
			fmt.Printf("this property is now:%v", property)
		}
		if propertySchema.Type == schema.TypeSet || propertySchema.Type == schema.TypeList {

			//(*localConfig)[name].(*schema.Set).List()
			if values, ok := responseConfig[property].(map[string]interface{}); ok {
				fmt.Printf("this property is now:%v", property)
				fmt.Printf("this property is now:%v", values)
				// configResult[property] = xInterfaceStrXStr(values.(*schema.Set).List())
				continue
			}

			// Na ovo obrati paznju
			// if v := c["apps"].(*schema.Set).List(); len(v) > 0 {
			// 	fivetranConfig.Apps(xInterfaceStrXStr(v))
			// }

			// if values := responseConfig[property].(*schema.Set).List(); len(values) > 0 {
			// 	fmt.Printf("this property is now:%v", property)
			// 	// configResult[property] = xInterfaceStrXStr(values.(*schema.Set).List())
			// 	continue
			// }

			// if v := c["adobe_analytics_configurations"].(*schema.Set).List(); len(v) > 0 {
			// 	fivetranConfig.AdobeAnalyticsConfigurations(resourceConnectorCreateConfigAdobeAnalyticsConfigurations(v))
			// }

			if values := responseConfig[property].(*schema.Set).List(); len(values) > 0 {
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
					if property == "custom_tables" {
						// if breakdownsValues := mapValues["breakdowns"].(*schema.Set).List(); len(breakdownsValues) > 0 {
						// 	fmt.Printf("this breakdownsValues is now:%v", breakdownsValues)
						// }
					}
					values[0] = mapValues
					configResult[property] = values
				} else {
					configResult[property] = xInterfaceStrXStr(values)
				}
				continue
			}
			if values, ok := responseConfig[property].(*schema.Set); ok {
				setValues := values.List()

				fmt.Printf("this property is now:%v", setValues)
			}

			if values, ok := responseConfig[property].([]string); ok {
				configResult[property] = xStrXInterface(values)
				continue
			}
			if interfaceValues, ok := responseConfig[property].([]interface{}); ok {
				if _, ok := interfaceValues[0].(map[string]interface{}); ok {
					configResult[property] = interfaceValues
				} else {
					configResult[property] = xInterfaceStrXStr(interfaceValues)
				}
				continue
			}
		}
		if value, ok := responseConfig[property].(string); ok && value != "" {
			valueType := propertySchema.Type
			switch valueType {
			case schema.TypeBool:
				configResult[property] = strToBool(value)
			case schema.TypeInt:
				configResult[property] = strToInt(value)
			default:
				configResult[property] = value
			}
			continue
		}
		if value, ok := responseConfig[property].(bool); ok {
			configResult[property] = value
			continue
		}
		if value, ok := responseConfig[property].(int); ok {
			configResult[property] = value
			continue
		}
	}

	return &configResult
}

func resourceConnectorAutomaticUpdateCustomAuth(d *schema.ResourceData) *map[string]interface{} {
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

func resourceConnectorAutomaticCreateConfig(fivetranConfig *fivetran.ConnectorConfig, destination_schema []interface{}) *fivetran.ConnectorConfig {
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

func resourceConnectorAutomaticCreateFunctionSecrets(xi []interface{}) []*fivetran.FunctionSecret {
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

func resourceConnectorAutomaticCreateConfigProjectCredentials(xi []interface{}) []*fivetran.ConnectorConfigProjectCredentials {
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

func resourceConnectorAutomaticCreateConfigCustomTables(xi []interface{}) []*fivetran.ConnectorConfigCustomTables {
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

func resourceConnectorAutomaticCreateConfigAdobeAnalyticsConfigurations(xi []interface{}) []*fivetran.ConnectorConfigAdobeAnalyticsConfiguration {
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

func resourceConnectorAutomaticCreateConfigReports(xi []interface{}) []*fivetran.ConnectorConfigReports {
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

func resourceConnectorAutomaticCreateAuth(auth []interface{}) *fivetran.ConnectorAuth {
	fivetranAuth := fivetran.NewConnectorAuth()

	if len(auth) < 1 {
		return fivetranAuth
	}
	if auth[0] == nil {
		return fivetranAuth
	}

	a := auth[0].(map[string]interface{})

	if v := a["client_access"].([]interface{}); len(v) > 0 {
		fivetranAuth.ClientAccess(resourceConnectorAutomaticCreateAuthClientAccess(v))
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

func resourceConnectorAutomaticCreateAuthClientAccess(clientAccess []interface{}) *fivetran.ConnectorAuthClientAccess {
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

// resourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func resourceConnectorAutomaticReadStatus(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", resourceConnectorAutomaticReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", resourceConnectorAutomaticReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

func resourceConnectorAutomaticReadStatusFlattenTasks(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Status.Tasks) < 1 {
		return make([]interface{}, 0)
	}

	tasks := make([]interface{}, len(resp.Data.Status.Tasks))
	for i, v := range resp.Data.Status.Tasks {
		task := make(map[string]interface{})
		mapAddStr(task, "code", v.Code)
		mapAddStr(task, "message", v.Message)

		tasks[i] = task
	}

	return tasks
}

func resourceConnectorAutomaticReadStatusFlattenWarnings(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Status.Warnings) < 1 {
		return make([]interface{}, 0)
	}

	warnings := make([]interface{}, len(resp.Data.Status.Warnings))
	for i, v := range resp.Data.Status.Warnings {
		warning := make(map[string]interface{})
		mapAddStr(warning, "code", v.Code)
		mapAddStr(warning, "message", v.Message)

		warnings[i] = warning
	}

	return warnings
}

// dataSourceConnectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func resourceConnectorAutomaticReadConfig(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig []interface{}) []interface{} {
	config := make([]interface{}, 1)

	configMap := make(map[string]interface{})

	responseConfig := resp.Data.CustomConfig

	responseConfigFromStruct := structToMap(resp.Data.Config)
	for responseProperty, value := range responseConfigFromStruct {
		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			var valueArray []interface{}
			for i := 0; i < reflectedValue.Len(); i++ {
				valueArray = append(valueArray, reflectedValue.Index(i).Interface())
			}

			if len(valueArray) > 0 {
				childPropertiesFromStruct := structToMap(valueArray[0])
				valueArray[0] = childPropertiesFromStruct
			}

			responseConfig[responseProperty] = valueArray
			continue
		}
		responseConfig[responseProperty] = value
	}

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getCSchemaAndProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			properties[key] = value
		}
	}

	// Ovde je BUG i problem za int i obj array's
	for property, propertySchema := range properties {
		if property == "is_ftps" {
			fmt.Printf(property)
		}
		if propertySchema.Type == schema.TypeSet {
			if values, ok := responseConfig[property].([]string); ok {
				configMap[property] = xStrXInterface(values)
				continue
			}
			if interfaceValues, ok := responseConfig[property].([]interface{}); ok {
				if len(interfaceValues) > 0 {
					if _, ok := interfaceValues[0].(map[string]interface{}); ok {
						configMap[property] = interfaceValues
					}
					continue
				}
				configMap[property] = xInterfaceStrXStr(interfaceValues)

				continue
			}

			continue
		}
		if v, ok := responseConfig[property].(string); ok && v != "" {
			valueType := propertySchema.Type
			switch valueType {
			case schema.TypeBool:
				configMap[property] = strToBool(v)
			case schema.TypeInt:
				configMap[property] = strToInt(v)
			default:
				configMap[property] = v
			}
		}
	}

	config[0] = configMap

	return config
}

func resourceConnectorAutomaticReadConfigFlattenProjectCredentials(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig []interface{}) []interface{} {
	if len(resp.Data.Config.ProjectCredentials) < 1 {
		return make([]interface{}, 0)
	}

	projectCredentials := make([]interface{}, len(resp.Data.Config.ProjectCredentials))
	for i, v := range resp.Data.Config.ProjectCredentials {
		pc := make(map[string]interface{})
		mapAddStr(pc, "project", v.Project)
		if len(currentConfig) > 0 {
			// The REST API sends the fields "api_key" and "secret_key" masked. We use the state stored config here.
			mapAddStr(pc, "api_key", resourceConnectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "api_key", currentConfig))
			mapAddStr(pc, "secret_key", resourceConnectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "secret_key", currentConfig))
		} else {
			// On Import these values will be masked, but we can't rely on state
			mapAddStr(pc, "api_key", v.APIKey)
			mapAddStr(pc, "secret_key", v.SecretKey)
		}
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func resourceConnectorAutomaticReadConfigFlattenSecretsList(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig []interface{}) []interface{} {
	if len(resp.Data.Config.SecretsList) < 1 {
		return make([]interface{}, 0)
	}
	secretsList := make([]interface{}, len(resp.Data.Config.SecretsList))
	for i, v := range resp.Data.Config.SecretsList {
		s := make(map[string]interface{})
		mapAddStr(s, "key", v.Key)
		if len(currentConfig) > 0 {
			mapAddStr(s, "value", resourceConnectorReadConfigFlattenSecretsListGetStateValue(v.Key, currentConfig))
		} else {
			mapAddStr(s, "value", v.Value)
		}
		secretsList[i] = s
	}

	return secretsList
}

func resourceConnectorAutomaticReadConfigFlattenProjectCredentialsGetStateValue(project, key string, currentConfig []interface{}) string {
	result := getSubcollectionElementValue("project_credentials", "project", project, key, currentConfig)

	if result == nil {
		return ""
	}

	return result.(string)
}

func resourceConnectorAutomaticReadConfigFlattenSecretsListGetStateValue(key string, currentConfig []interface{}) string {
	result := getSubcollectionElementValue("secrets_list", "key", key, "value", currentConfig)

	if result == nil {
		return ""
	}

	return result.(string)
}

func getSubcollectionAutomaticElementValue(configKey, subKey, subKeyValue, targetKey string, currentConfig []interface{}) interface{} {
	targetList := currentConfig[0].(map[string]interface{})[configKey].(*schema.Set).List()
	for _, v := range targetList {
		if v.(map[string]interface{})[subKey].(string) == subKeyValue {
			return v.(map[string]interface{})[targetKey]
		}
	}
	return nil
}

func resourceConnectorAutomaticReadConfigFlattenReports(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Config.Reports) < 1 {
		return make([]interface{}, 0)
	}

	reports := make([]interface{}, len(resp.Data.Config.Reports))
	for i, v := range resp.Data.Config.Reports {
		r := make(map[string]interface{})
		mapAddStr(r, "table", v.Table)
		mapAddStr(r, "config_type", v.ConfigType)
		mapAddStr(r, "prebuilt_report", v.PrebuiltReport)
		mapAddStr(r, "report_type", v.ReportType)
		mapAddXInterface(r, "fields", xStrXInterface(v.Fields))
		mapAddXInterface(r, "dimensions", xStrXInterface(v.Dimensions))
		mapAddXInterface(r, "metrics", xStrXInterface(v.Metrics))
		mapAddXInterface(r, "segments", xStrXInterface(v.Segments))
		mapAddStr(r, "filter", v.Filter)
		reports[i] = r
	}

	return reports
}

func resourceConnectorAutomaticReadConfigFlattenAdobeAnalyticsConfigurations(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Config.AdobeAnalyticsConfigurations) < 1 {
		return make([]interface{}, 0)
	}

	configurations := make([]interface{}, len(resp.Data.Config.AdobeAnalyticsConfigurations))
	for i, v := range resp.Data.Config.AdobeAnalyticsConfigurations {
		c := make(map[string]interface{})
		mapAddStr(c, "sync_mode", v.SyncMode)
		mapAddXInterface(c, "metrics", xStrXInterface(v.Metrics))
		mapAddXInterface(c, "calculated_metrics", xStrXInterface(v.CalculatedMetrics))
		mapAddXInterface(c, "elements", xStrXInterface(v.Elements))
		mapAddXInterface(c, "segments", xStrXInterface(v.Segments))
		mapAddXInterface(c, "report_suites", xStrXInterface(v.ReportSuites))
		configurations[i] = c
	}

	return configurations
}

func resourceConnectorAutomaticReadConfigFlattenCustomTables(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Config.CustomTables) < 1 {
		return make([]interface{}, 0)
	}

	customTables := make([]interface{}, len(resp.Data.Config.CustomTables))
	for i, v := range resp.Data.Config.CustomTables {
		ct := make(map[string]interface{})
		mapAddStr(ct, "table_name", v.TableName)
		mapAddStr(ct, "config_type", v.ConfigType)
		mapAddXInterface(ct, "fields", xStrXInterface(v.Fields))
		mapAddXInterface(ct, "breakdowns", xStrXInterface(v.Breakdowns))
		mapAddXInterface(ct, "action_breakdowns", xStrXInterface(v.ActionBreakdowns))
		mapAddStr(ct, "aggregation", v.Aggregation)
		mapAddStr(ct, "action_report_time", v.ActionReportTime)
		mapAddStr(ct, "click_attribution_window", v.ClickAttributionWindow)
		mapAddStr(ct, "view_attribution_window", v.ViewAttributionWindow)
		mapAddStr(ct, "prebuilt_report_name", v.PrebuiltReportName)
		customTables[i] = ct
	}

	return customTables
}
