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
		Schema:        connectorSchema(false),
	}
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
	svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))
	svc.Paused(strToBool(d.Get("paused").(string)))
	svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	if d.Get("sync_frequency") == "1440" && d.Get("daily_sync_time").(string) != "" {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	fivetranConfig := resourceConnectorUpdateConfig(d)

	svc.Config(resourceConnectorCreateConfig(fivetranConfig, d.Get("destination_schema").([]interface{})))
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

	msi := connectorRead(&currentConfig, resp)

	currentService := d.Get("service").(string)

	// Ignore service change for migrated `adwords` connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(msi, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	}

	// Value for daily_sync_time won't be returned if sync_frequency < 1440
	if *resp.Data.SyncFrequency != 1440 {
		mapAddStr(msi, "daily_sync_time", d.Get("daily_sync_time").(string))
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

	svc.Config(resourceConnectorUpdateConfig(d))
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
	configMap := make(map[string]interface{})

	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return &configMap
	}
	if config[0] == nil {
		return &configMap
	}

	c := config[0].(map[string]interface{})

	if v, ok := c["group_name"].(string); ok && v != "" {
		configMap["group_name"] = v
	}

	if v, ok := c["sync_method"].(string); ok && v != "" {
		configMap["sync_method"] = v
	}

	if v, ok := c["is_account_level_connector"].(string); ok && v != "" {
		configMap["is_account_level_connector"] = strToBool(v)
	}

	// HVA connector parameters

	if v, ok := c["pdb_name"].(string); ok && v != "" {
		configMap["pdb_name"] = v
	}

	if v, ok := c["agent_host"].(string); ok && v != "" {
		configMap["agent_host"] = v
	}

	if v, ok := c["agent_port"].(string); ok && v != "" {
		configMap["agent_port"] = strToInt(v)
	}

	if v, ok := c["agent_user"].(string); ok && v != "" {
		configMap["agent_user"] = v
	}

	if v, ok := c["agent_password"].(string); ok && v != "" {
		configMap["agent_password"] = v
	}

	if v, ok := c["agent_public_cert"].(string); ok && v != "" {
		configMap["agent_public_cert"] = v
	}

	if v, ok := c["agent_ora_home"].(string); ok && v != "" {
		configMap["agent_ora_home"] = v
	}

	if v, ok := c["tns"].(string); ok && v != "" {
		configMap["tns"] = v
	}

	if v, ok := c["use_oracle_rac"].(string); ok && v != "" {
		configMap["use_oracle_rac"] = strToBool(v)
	}

	if v, ok := c["asm_option"].(string); ok && v != "" {
		configMap["asm_option"] = strToBool(v)
	}

	if v, ok := c["asm_user"].(string); ok && v != "" {
		configMap["asm_user"] = v
	}

	if v, ok := c["asm_password"].(string); ok && v != "" {
		configMap["asm_password"] = v
	}

	if v, ok := c["asm_oracle_home"].(string); ok && v != "" {
		configMap["asm_oracle_home"] = v
	}

	if v, ok := c["asm_tns"].(string); ok && v != "" {
		configMap["asm_tns"] = v
	}

	if v, ok := c["sap_user"].(string); ok && v != "" {
		configMap["sap_user"] = v
	}

	if v, ok := c["organization"].(string); ok && v != "" {
		configMap["organization"] = v
	}

	if v, ok := c["packed_mode_tables"]; ok {
		configMap["packed_mode_tables"] = xInterfaceStrXStr(v.(*schema.Set).List())
	}

	if v, ok := c["access_key"].(string); ok && v != "" {
		configMap["access_key"] = v
	}

	if v, ok := c["domain_host_name"].(string); ok && v != "" {
		configMap["domain_host_name"] = v
	}

	if v, ok := c["client_name"].(string); ok && v != "" {
		configMap["client_name"] = v
	}

	if v, ok := c["domain_type"].(string); ok && v != "" {
		configMap["domain_type"] = v
	}

	if v, ok := c["connection_method"].(string); ok && v != "" {
		configMap["connection_method"] = v
	}

	if v, ok := c["is_single_table_mode"].(string); ok && v != "" {
		configMap["is_single_table_mode"] = strToBool(v)
	}

	if v, ok := c["company_id"].(string); ok && v != "" {
		configMap["company_id"] = v
	}

	if v, ok := c["login_password"].(string); ok && v != "" {
		configMap["login_password"] = v
	}

	if v, ok := c["environment"].(string); ok && v != "" {
		configMap["environment"] = v
	}

	if v, ok := c["properties"]; ok {
		configMap["properties"] = xInterfaceStrXStr(v.(*schema.Set).List())
	}

	if v, ok := c["is_public"].(string); ok && v != "" {
		configMap["is_public"] = strToBool(v)
	}

	if v, ok := c["empty_header"].(string); ok && v != "" {
		configMap["empty_header"] = strToBool(v)
	}

	if v, ok := c["list_strategy"].(string); ok && v != "" {
		configMap["list_strategy"] = v
	}

	if v, ok := c["support_nested_columns"].(string); ok && v != "" {
		configMap["support_nested_columns"] = strToBool(v)
	}

	if v, ok := c["csv_definition"].(string); ok && v != "" {
		configMap["csv_definition"] = v
	}

	if v, ok := c["export_storage_type"].(string); ok && v != "" {
		configMap["export_storage_type"] = v
	}

	if v, ok := c["primary_keys"]; ok {
		configMap["primary_keys"] = xInterfaceStrXStr(v.(*schema.Set).List())
	}

	// HVA parameters end

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

	// add custom auth fields here:

	// a := auth[0].(map[string]interface{})

	// if v := a["some_auth_custom_field"].(string); v != "" {
	// 	authMap["some_auth_custom_field"] = v
	// }

	return &authMap
}

func resourceConnectorUpdateConfig(d *schema.ResourceData) *fivetran.ConnectorConfig {
	fivetranConfig := fivetran.NewConnectorConfig()
	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return fivetranConfig
	}
	if config[0] == nil {
		return fivetranConfig
	}

	c := config[0].(map[string]interface{})

	if v := c["sheet_id"].(string); v != "" {
		fivetranConfig.SheetID(v)
	}
	if v := c["share_url"].(string); v != "" {
		fivetranConfig.ShareURL(v)
	}
	if v := c["named_range"].(string); v != "" {
		fivetranConfig.NamedRange(v)
	}
	if v := c["client_id"].(string); v != "" {
		fivetranConfig.ClientID(v)
	}
	if v := c["client_secret"].(string); v != "" {
		fivetranConfig.ClientSecret(v)
	}
	if v := c["technical_account_id"].(string); v != "" {
		fivetranConfig.TechnicalAccountID(v)
	}
	if v := c["organization_id"].(string); v != "" {
		fivetranConfig.OrganizationID(v)
	}
	if v := c["private_key"].(string); v != "" {
		fivetranConfig.PrivateKey(v)
	}
	if v := c["sync_mode"].(string); v != "" {
		fivetranConfig.SyncMode(v)
	}
	if v := c["report_suites"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.ReportSuites(xInterfaceStrXStr(v))
	}
	if v := c["elements"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Elements(xInterfaceStrXStr(v))
	}
	if v := c["metrics"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Metrics(xInterfaceStrXStr(v))
	}
	if v := c["date_granularity"].(string); v != "" {
		fivetranConfig.DateGranularity(v)
	}
	if v := c["timeframe_months"].(string); v != "" {
		fivetranConfig.TimeframeMonths(v)
	}
	if v := c["source"].(string); v != "" {
		fivetranConfig.Source(v)
	}
	if v := c["s3bucket"].(string); v != "" {
		fivetranConfig.S3Bucket(v)
	}
	if v := c["s3role_arn"].(string); v != "" {
		fivetranConfig.S3RoleArn(v)
	}
	if v := c["abs_connection_string"].(string); v != "" {
		fivetranConfig.ABSConnectionString(v)
	}
	if v := c["abs_container_name"].(string); v != "" {
		fivetranConfig.ABSContainerName(v)
	}
	if v := c["folder_id"].(string); v != "" {
		fivetranConfig.FolderId(v)
	}
	if v := c["ftp_host"].(string); v != "" {
		fivetranConfig.FTPHost(v)
	}
	if v := c["ftp_port"].(string); v != "" {
		fivetranConfig.FTPPort(strToInt(v))
	}
	if v := c["ftp_user"].(string); v != "" {
		fivetranConfig.FTPUser(v)
	}
	if v := c["ftp_password"].(string); v != "" {
		fivetranConfig.FTPPassword(v)
	}
	if v := c["is_ftps"].(string); v != "" {
		fivetranConfig.IsFTPS(strToBool(v))
	}
	if v := c["sftp_host"].(string); v != "" {
		fivetranConfig.SFTPHost(v)
	}
	if v := c["sftp_port"].(string); v != "" {
		fivetranConfig.SFTPPort(strToInt(v))
	}
	if v := c["sftp_user"].(string); v != "" {
		fivetranConfig.SFTPUser(v)
	}
	if v := c["sftp_password"].(string); v != "" {
		fivetranConfig.SFTPPassword(v)
	}
	if v := c["sftp_is_key_pair"].(string); v != "" {
		fivetranConfig.SFTPIsKeyPair(strToBool(v))
	}
	if v := c["is_keypair"].(string); v != "" {
		fivetranConfig.IsKeypair(strToBool(v))
	}
	if v := c["advertisables"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Advertisables(xInterfaceStrXStr(v))
	}
	if v := c["report_type"].(string); v != "" {
		fivetranConfig.ReportType(v)
	}
	if v := c["dimensions"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Dimensions(xInterfaceStrXStr(v))
	}
	if v := c["api_key"].(string); v != "" {
		fivetranConfig.APIKey(v)
	}
	if v := c["external_id"].(string); v != "" {
		fivetranConfig.ExternalID(v)
	}
	if v := c["role_arn"].(string); v != "" {
		fivetranConfig.RoleArn(v)
	}
	if v := c["bucket"].(string); v != "" {
		fivetranConfig.Bucket(v)
	}
	if v := c["prefix"].(string); v != "" {
		fivetranConfig.Prefix(v)
	}
	if v := c["pattern"].(string); v != "" {
		fivetranConfig.Pattern(v)
	}
	if v := c["file_type"].(string); v != "" {
		fivetranConfig.FileType(v)
	}
	if v := c["compression"].(string); v != "" {
		fivetranConfig.Compression(v)
	}
	if v := c["on_error"].(string); v != "" {
		fivetranConfig.OnError(v)
	}
	if v := c["append_file_option"].(string); v != "" {
		fivetranConfig.AppendFileOption(v)
	}
	if v := c["archive_pattern"].(string); v != "" {
		fivetranConfig.ArchivePattern(v)
	}
	if v := c["null_sequence"].(string); v != "" {
		fivetranConfig.NullSequence(v)
	}
	if v := c["delimiter"].(string); v != "" {
		fivetranConfig.Delimiter(v)
	}
	if v := c["escape_char"].(string); v != "" {
		fivetranConfig.EscapeChar(v)
	}
	if v := c["skip_before"].(string); v != "" {
		fivetranConfig.SkipBefore(strToInt(v))
	}
	if v := c["skip_after"].(string); v != "" {
		fivetranConfig.SkipAfter(strToInt(v))
	}
	if v := c["project_credentials"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.ProjectCredentials(resourceConnectorCreateConfigProjectCredentials(v))
	}
	if v := c["secrets_list"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.SecretsList(resourceConnectorCreateFunctionSecrets(v))
	}
	if v := c["auth_mode"].(string); v != "" {
		fivetranConfig.AuthMode(v)
	}
	if v := c["username"].(string); v != "" {
		fivetranConfig.Username(v)
	}
	if v := c["user_name"].(string); v != "" {
		fivetranConfig.UserName(v)
	}
	if v := c["password"].(string); v != "" {
		fivetranConfig.Password(v)
	}
	if v := c["certificate"].(string); v != "" {
		fivetranConfig.Certificate(v)
	}
	if v := c["selected_exports"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.SelectedExports(xInterfaceStrXStr(v))
	}
	if v := c["consumer_group"].(string); v != "" {
		fivetranConfig.ConsumerGroup(v)
	}
	if v := c["servers"].(string); v != "" {
		fivetranConfig.Servers(v)
	}
	if v := c["message_type"].(string); v != "" {
		fivetranConfig.MessageType(v)
	}
	if v := c["sync_type"].(string); v != "" {
		fivetranConfig.SyncType(v)
	}
	if v := c["security_protocol"].(string); v != "" {
		fivetranConfig.SecurityProtocol(v)
	}
	if v := c["apps"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Apps(xInterfaceStrXStr(v))
	}
	if v := c["sales_accounts"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.SalesAccounts(xInterfaceStrXStr(v))
	}
	if v := c["finance_accounts"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.FinanceAccounts(xInterfaceStrXStr(v))
	}
	if v := c["app_sync_mode"].(string); v != "" {
		fivetranConfig.AppSyncMode(v)
	}
	if v := c["sales_account_sync_mode"].(string); v != "" {
		fivetranConfig.SalesAccountSyncMode(v)
	}
	if v := c["finance_account_sync_mode"].(string); v != "" {
		fivetranConfig.FinanceAccountSyncMode(v)
	}
	if v := c["pem_certificate"].(string); v != "" {
		fivetranConfig.PEMCertificate(v)
	}
	if v := c["access_key_id"].(string); v != "" {
		fivetranConfig.AccessKeyID(v)
	}
	if v := c["secret_key"].(string); v != "" {
		fivetranConfig.SecretKey(v)
	}
	if v := c["home_folder"].(string); v != "" {
		fivetranConfig.HomeFolder(v)
	}
	if v := c["sync_data_locker"].(string); v != "" {
		fivetranConfig.SyncDataLocker(strToBool(v))
	}
	if v := c["projects"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Projects(xInterfaceStrXStr(v))
	}
	if v := c["function"].(string); v != "" {
		fivetranConfig.Function(v)
	}
	if v := c["region"].(string); v != "" {
		fivetranConfig.Region(v)
	}
	if v := c["secrets"].(string); v != "" {
		fivetranConfig.Secrets(v)
	}
	if v := c["container_name"].(string); v != "" {
		fivetranConfig.ContainerName(v)
	}
	if v := c["connection_string"].(string); v != "" {
		fivetranConfig.ConnectionString(v)
	}
	if v := c["function_app"].(string); v != "" {
		fivetranConfig.FunctionApp(v)
	}
	if v := c["function_name"].(string); v != "" {
		fivetranConfig.FunctionName(v)
	}
	if v := c["function_key"].(string); v != "" {
		fivetranConfig.FunctionKey(v)
	}
	if v := c["merchant_id"].(string); v != "" {
		fivetranConfig.MerchantID(v)
	}
	if v := c["api_url"].(string); v != "" {
		fivetranConfig.APIURL(v)
	}
	if v := c["cloud_storage_type"].(string); v != "" {
		fivetranConfig.CloudStorageType(v)
	}
	if v := c["s3external_id"].(string); v != "" {
		fivetranConfig.S3ExternalID(v)
	}
	if v := c["s3folder"].(string); v != "" {
		fivetranConfig.S3Folder(v)
	}
	if v := c["gcs_bucket"].(string); v != "" {
		fivetranConfig.GCSBucket(v)
	}
	if v := c["gcs_folder"].(string); v != "" {
		fivetranConfig.GCSFolder(v)
	}
	if v := c["user_profiles"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.UserProfiles(xInterfaceStrXStr(v))
	}
	if v := c["report_configuration_ids"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.ReportConfigurationIDs(xInterfaceStrXStr(v))
	}
	if v := c["enable_all_dimension_combinations"].(string); v != "" {
		fivetranConfig.EnableAllDimensionCombinations(strToBool(v))
	}
	if v := c["instance"].(string); v != "" {
		fivetranConfig.Instance(v)
	}
	if v := c["aws_region_code"].(string); v != "" {
		fivetranConfig.AWSRegionCode(v)
	}
	if v := c["accounts"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Accounts(xInterfaceStrXStr(v))
	}
	if v := c["fields"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Fields(xInterfaceStrXStr(v))
	}
	if v := c["breakdowns"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Breakdowns(xInterfaceStrXStr(v))
	}
	if v := c["action_breakdowns"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.ActionBreakdowns(xInterfaceStrXStr(v))
	}
	if v := c["aggregation"].(string); v != "" {
		fivetranConfig.Aggregation(v)
	}
	if v := c["config_type"].(string); v != "" {
		fivetranConfig.ConfigType(v)
	}
	if v := c["prebuilt_report"].(string); v != "" {
		fivetranConfig.PrebuiltReport(v)
	}
	if v := c["action_report_time"].(string); v != "" {
		fivetranConfig.ActionReportTime(v)
	}
	if v := c["click_attribution_window"].(string); v != "" {
		fivetranConfig.ClickAttributionWindow(v)
	}
	if v := c["view_attribution_window"].(string); v != "" {
		fivetranConfig.ViewAttributionWindow(v)
	}
	if v := c["custom_tables"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.CustomTables(resourceConnectorCreateConfigCustomTables(v))
	}
	if v := c["pages"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Pages(xInterfaceStrXStr(v))
	}
	if v := c["subdomain"].(string); v != "" {
		fivetranConfig.Subdomain(v)
	}
	if v := c["port"].(string); v != "" {
		fivetranConfig.Port(strToInt(v))
	}
	if v := c["user"].(string); v != "" {
		fivetranConfig.User(v)
	}
	if v := c["is_secure"].(string); v != "" {
		fivetranConfig.IsSecure(strToBool(v))
	}
	if v := c["repositories"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Repositories(xInterfaceStrXStr(v))
	}
	if v := c["use_webhooks"].(string); v != "" {
		fivetranConfig.UseWebhooks(strToBool(v))
	}
	if v := c["dimension_attributes"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.DimensionAttributes(xInterfaceStrXStr(v))
	}
	if v := c["columns"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Columns(xInterfaceStrXStr(v))
	}
	if v := c["network_code"].(string); v != "" {
		fivetranConfig.NetworkCode(v)
	}
	if v := c["customer_id"].(string); v != "" {
		fivetranConfig.CustomerID(v)
	}
	if v := c["manager_accounts"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.ManagerAccounts(xInterfaceStrXStr(v))
	}
	if v := c["reports"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Reports(resourceConnectorCreateConfigReports(v))
	}
	if v := c["conversion_window_size"].(string); v != "" {
		fivetranConfig.ConversionWindowSize(strToInt(v))
	}
	if v := c["profiles"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Profiles(xInterfaceStrXStr(v))
	}
	if v := c["project_id"].(string); v != "" {
		fivetranConfig.ProjectID(v)
	}
	if v := c["dataset_id"].(string); v != "" {
		fivetranConfig.DatasetID(v)
	}
	if v := c["bucket_name"].(string); v != "" {
		fivetranConfig.BucketName(v)
	}
	if v := c["function_trigger"].(string); v != "" {
		fivetranConfig.FunctionTrigger(v)
	}
	if v := c["config_method"].(string); v != "" {
		fivetranConfig.ConfigMethod(v)
	}
	if v := c["query_id"].(string); v != "" {
		fivetranConfig.QueryID(v)
	}
	if v := c["update_config_on_each_sync"].(string); v != "" {
		fivetranConfig.UpdateConfigOnEachSync(strToBool(v))
	}
	if v := c["site_urls"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.SiteURLs(xInterfaceStrXStr(v))
	}
	if v := c["path"].(string); v != "" {
		fivetranConfig.Path(v)
	}
	if v := c["on_premise"].(string); v != "" {
		fivetranConfig.OnPremise(strToBool(v))
	}
	if v := c["access_token"].(string); v != "" {
		fivetranConfig.AccessToken(v)
	}
	if v := c["view_through_attribution_window_size"].(string); v != "" {
		fivetranConfig.ViewThroughAttributionWindowSize(v)
	}
	if v := c["post_click_attribution_window_size"].(string); v != "" {
		fivetranConfig.PostClickAttributionWindowSize(v)
	}
	if v := c["use_api_keys"].(string); v != "" {
		fivetranConfig.UseAPIKeys(strToBool(v))
	}
	if v := c["api_keys"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.APIKeys(xInterfaceStrXStr(v))
	}
	if v := c["endpoint"].(string); v != "" {
		fivetranConfig.Endpoint(v)
	}
	if v := c["identity"].(string); v != "" {
		fivetranConfig.Identity(v)
	}
	if v := c["api_quota"].(string); v != "" {
		fivetranConfig.APIQuota(strToInt(v))
	}
	if v := c["domain_name"].(string); v != "" {
		fivetranConfig.DomainName(v)
	}
	if v := c["resource_url"].(string); v != "" {
		fivetranConfig.ResourceURL(v)
	}
	if v := c["api_secret"].(string); v != "" {
		fivetranConfig.APISecret(v)
	}
	if v := c["host"].(string); v != "" {
		fivetranConfig.Host(v)
	}
	if v := c["hosts"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Hosts(xInterfaceStrXStr(v))
	}
	if v := c["tunnel_host"].(string); v != "" {
		fivetranConfig.TunnelHost(v)
	}
	if v := c["tunnel_port"].(string); v != "" {
		fivetranConfig.TunnelPort(strToInt(v))
	}
	if v := c["tunnel_user"].(string); v != "" {
		fivetranConfig.TunnelUser(v)
	}
	if v := c["database"].(string); v != "" {
		fivetranConfig.Database(v)
	}
	if v := c["datasource"].(string); v != "" {
		fivetranConfig.Datasource(v)
	}
	if v := c["account"].(string); v != "" {
		fivetranConfig.Account(v)
	}
	if v := c["role"].(string); v != "" {
		fivetranConfig.Role(v)
	}
	if v := c["email"].(string); v != "" {
		fivetranConfig.Email(v)
	}
	if v := c["account_id"].(string); v != "" {
		fivetranConfig.AccountID(v)
	}
	if v := c["server_url"].(string); v != "" {
		fivetranConfig.ServerURL(v)
	}
	if v := c["user_key"].(string); v != "" {
		fivetranConfig.UserKey(v)
	}
	if v := c["api_version"].(string); v != "" {
		fivetranConfig.APIVersion(v)
	}
	if v := c["daily_api_call_limit"].(string); v != "" {
		fivetranConfig.DailyAPICallLimit(strToInt(v))
	}
	if v := c["time_zone"].(string); v != "" {
		fivetranConfig.TimeZone(v)
	}
	if v := c["integration_key"].(string); v != "" {
		fivetranConfig.IntegrationKey(v)
	}
	if v := c["advertisers"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Advertisers(xInterfaceStrXStr(v))
	}
	if v := c["engagement_attribution_window"].(string); v != "" {
		fivetranConfig.EngagementAttributionWindow(v)
	}
	if v := c["conversion_report_time"].(string); v != "" {
		fivetranConfig.ConversionReportTime(v)
	}
	if v := c["domain"].(string); v != "" {
		fivetranConfig.Domain(v)
	}
	if v := c["update_method"].(string); v != "" {
		fivetranConfig.UpdateMethod(v)
	}
	if v := c["connection_type"].(string); v != "" {
		fivetranConfig.ConnectionType(v)
	}
	if v := c["replication_slot"].(string); v != "" {
		fivetranConfig.ReplicationSlot(v)
	}
	if v := c["publication_name"].(string); v != "" {
		fivetranConfig.PublicationName(v)
	}
	if v := c["data_center"].(string); v != "" {
		fivetranConfig.DataCenter(v)
	}
	if v := c["api_token"].(string); v != "" {
		fivetranConfig.APIToken(v)
	}
	if v := c["sub_domain"].(string); v != "" {
		fivetranConfig.SubDomain(v)
	}
	if v := c["test_table_name"].(string); v != "" {
		fivetranConfig.TestTableName(v)
	}
	if v := c["shop"].(string); v != "" {
		fivetranConfig.Shop(v)
	}
	if v := c["organizations"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.Organizations(xInterfaceStrXStr(v))
	}
	if v := c["swipe_attribution_window"].(string); v != "" {
		fivetranConfig.SwipeAttributionWindow(v)
	}
	if v := c["api_access_token"].(string); v != "" {
		fivetranConfig.APIAccessToken(v)
	}
	if v := c["account_ids"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.AccountIDs(xInterfaceStrXStr(v))
	}
	if v := c["sid"].(string); v != "" {
		fivetranConfig.SID(v)
	}
	if v := c["secret"].(string); v != "" {
		fivetranConfig.Secret(v)
	}
	if v := c["oauth_token"].(string); v != "" {
		fivetranConfig.OauthToken(v)
	}
	if v := c["oauth_token_secret"].(string); v != "" {
		fivetranConfig.OauthTokenSecret(v)
	}
	if v := c["consumer_key"].(string); v != "" {
		fivetranConfig.ConsumerKey(v)
	}
	if v := c["consumer_secret"].(string); v != "" {
		fivetranConfig.ConsumerSecret(v)
	}
	if v := c["key"].(string); v != "" {
		fivetranConfig.Key(v)
	}
	if v := c["advertisers_id"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.AdvertisersID(xInterfaceStrXStr(v))
	}
	if v := c["sync_format"].(string); v != "" {
		fivetranConfig.SyncFormat(v)
	}
	if v := c["bucket_service"].(string); v != "" {
		fivetranConfig.BucketService(v)
	}

	if v := c["report_url"].(string); v != "" {
		fivetranConfig.ReportURL(v)
	}
	if v := c["unique_id"].(string); v != "" {
		fivetranConfig.UniqueID(v)
	}
	if v := c["auth_type"].(string); v != "" {
		fivetranConfig.AuthType(v)
	}
	if v := c["is_new_package"].(string); v != "" {
		fivetranConfig.IsNewPackage(strToBool(v))
	}
	if v := c["adobe_analytics_configurations"].(*schema.Set).List(); len(v) > 0 {
		fivetranConfig.AdobeAnalyticsConfigurations(resourceConnectorCreateConfigAdobeAnalyticsConfigurations(v))
	}
	if v := c["is_multi_entity_feature_enabled"].(string); v != "" {
		fivetranConfig.IsMultiEntityFeatureEnabled(strToBool(v))
	}
	if v := c["api_type"].(string); v != "" {
		fivetranConfig.ApiType(v)
	}
	if v := c["base_url"].(string); v != "" {
		fivetranConfig.BaseUrl(v)
	}
	if v := c["entity_id"].(string); v != "" {
		fivetranConfig.EntityId(v)
	}
	if v := c["soap_uri"].(string); v != "" {
		fivetranConfig.SoapUri(v)
	}
	if v := c["user_id"].(string); v != "" {
		fivetranConfig.UserId(v)
	}
	if v := c["encryption_key"].(string); v != "" {
		fivetranConfig.EncryptionKey(v)
	}
	if v := c["pat"].(string); v != "" {
		fivetranConfig.PAT(v)
	}
	if v := c["always_encrypted"].(string); v != "" {
		fivetranConfig.AlwaysEncrypted(strToBool(v))
	}
	if v := c["eu_region"].(string); v != "" {
		fivetranConfig.EuRegion(strToBool(v))
	}
	if v := c["token_key"].(string); v != "" {
		fivetranConfig.TokenKey(v)
	}
	if v := c["token_secret"].(string); v != "" {
		fivetranConfig.TokenSecret(v)
	}
	if v := c["public_key"].(string); v != "" {
		fivetranConfig.PublicKey(v)
	}

	return fivetranConfig
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
