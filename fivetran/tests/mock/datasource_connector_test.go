package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	connectorDataSourceMockGetHandler *mock.Handler

	connectorDataSourceMockData map[string]interface{}
)

func setupMockClientConnectorDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	connectorDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorDataSourceMockData = createMapFromJsonString(t, connectorMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorDataSourceMockData), nil
		},
	)
}

func TestDataSourceConnectorConfigMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connector" "test_connector" {
			provider = fivetran-provider
			id = "connector_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, connectorDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service", "google_sheets"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service", "google_sheets"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service_version", "1"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "schedule_type", "auto"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.is_historical_sync", "true"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.update_state", "on_schedule"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.setup_state", "incomplete"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.sync_state", "paused"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.tasks.0.code", "task_code"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.tasks.0.message", "task_message"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.warnings.0.code", "warning_code"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.0.warnings.0.message", "warning_message"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "name", "google_sheets_schema.table"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "sync_frequency", "5"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "paused", "true"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "pause_after_trial", "true"),

			// check sensitive fields are have original values
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.oauth_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.oauth_token_secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.consumer_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.client_secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.private_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.s3role_arn", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.ftp_password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.sftp_password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.api_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.role_arn", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.secret_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.pem_certificate", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.access_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.api_secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.api_access_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.consumer_secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.secrets", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.api_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.encryption_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.pat", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.function_trigger", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.token_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.token_secret", "******"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.sheet_id", "sheet_id"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.auth_type", "OAuth"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.named_range", "range"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.sync_method", "sync_method"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.is_ftps", "false"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.sftp_is_key_pair", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.sync_data_locker", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.enable_all_dimension_combinations", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.update_config_on_each_sync", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.on_premise", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.is_new_package", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.is_multi_entity_feature_enabled", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.always_encrypted", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.is_secure", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.use_api_keys", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.use_webhooks", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.eu_region", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.0.is_keypair", "false"),

			// conversion_window_size = "0"
			// skip_before = "0"
			// skip_after = "0"
			// ftp_port = "0"
			// sftp_port = "0"
			// port = "0"
			// tunnel_port = "0"
			// api_quota = "0"
			// daily_api_call_limit = "0"

			// connection_type = "connection_type"
			// sync_mode = "sync_mode"
			// date_granularity = "date_granularity"
			// timeframe_months = "timeframe_months"
			// report_type = "report_type"
			// aggregation = "aggregation"
			// config_type = "config_type"
			// prebuilt_report = "prebuilt_report"
			// action_report_time = "action_report_time"
			// click_attribution_window = "click_attribution_window"
			// view_attribution_window = "view_attribution_window"
			// view_through_attribution_window_size = "view_through_attribution_window_size"
			// post_click_attribution_window_size = "post_click_attribution_window_size"
			// update_method = "update_method"
			// swipe_attribution_window = "swipe_attribution_window"
			// api_type = "api_type"
			// sync_format = "sync_format"
			// app_sync_mode = "app_sync_mode"
			// sales_account_sync_mode = "sales_account_sync_mode"
			// finance_account_sync_mode = "finance_account_sync_mode"
			// source = "source"
			// file_type = "file_type"
			// compression = "compression"
			// on_error = "on_error"
			// append_file_option = "append_file_option"
			// engagement_attribution_window = "engagement_attribution_window"
			// conversion_report_time = "conversion_report_time"

			// external_id = "external_id"
			// public_key = "public_key"

			// client_id = "client_id"
			// technical_account_id = "technical_account_id"
			// organization_id = "organization_id"
			// s3bucket = "s3bucket"
			// abs_connection_string = "abs_connection_string"
			// abs_container_name = "abs_container_name"
			// folder_id = "folder_id"
			// ftp_host = "ftp_host"
			// ftp_user = "ftp_user"
			// sftp_host = "sftp_host"
			// sftp_user = "sftp_user"
			// bucket = "bucket"
			// prefix = "prefix"
			// pattern = "pattern"
			// archive_pattern = "archive_pattern"
			// null_sequence = "null_sequence"
			// delimiter = "delimiter"
			// escape_char = "escape_char"
			// auth_mode = "auth_mode"
			// certificate = "certificate"
			// consumer_group = "consumer_group"
			// servers = "servers"
			// message_type = "message_type"
			// sync_type = "sync_type"
			// security_protocol = "security_protocol"
			// access_key_id = "access_key_id"
			// home_folder = "home_folder"
			// function = "function"
			// region = "region"
			// container_name = "container_name"
			// connection_string = "connection_string"
			// function_app = "function_app"
			// function_name = "function_name"
			// function_key = "function_key"
			// merchant_id = "merchant_id"
			// api_url = "api_url"
			// cloud_storage_type = "cloud_storage_type"
			// s3external_id = "s3external_id"
			// s3folder = "s3folder"
			// gcs_bucket = "gcs_bucket"
			// gcs_folder = "gcs_folder"
			// instance = "instance"
			// aws_region_code = "aws_region_code"
			// host = "host"

			// user = "user"
			// network_code = "network_code"
			// customer_id = "customer_id"
			// project_id = "project_id"
			// dataset_id = "dataset_id"
			// bucket_name = "bucket_name"
			// config_method = "config_method"
			// query_id = "query_id"
			// path = "path"
			// endpoint = "endpoint"
			// identity = "identity"

			// domain_name = "domain_name"
			// resource_url = "resource_url"
			// tunnel_host = "tunnel_host"
			// tunnel_user = "tunnel_user"
			// database = "database"
			// datasource = "datasource"
			// account = "account"
			// role = "role"
			// email = "email"
			// account_id = "account_id"
			// server_url = "server_url"
			// user_key = "user_key"
			// api_version = "api_version"
			// time_zone = "time_zone"
			// integration_key = "integration_key"
			// domain = "domain"
			// replication_slot = "replication_slot"
			// publication_name = "publication_name"
			// data_center = "data_center"
			// sub_domain = "sub_domain"
			// subdomain = "subdomain"
			// test_table_name = "test_table_name"
			// shop = "shop"
			// sid = "sid"
			// key = "key"
			// bucket_service = "bucket_service"
			// user_name = "user_name"
			// username = "username"
			// report_url = "report_url"
			// unique_id = "unique_id"
			// base_url = "base_url"
			// entity_id = "entity_id"
			// soap_uri = "soap_uri"
			// user_id = "user_id"
			// share_url = "share_url"

			// report_suites = ["report_suite"]
			// elements = ["element"]
			// metrics = ["metric"]
			// advertisables = ["advertisable"]
			// dimensions = ["dimension"]
			// selected_exports = ["selected_export"]
			// apps = ["app"]
			// sales_accounts = ["sales_account"]
			// finance_accounts = ["finance_account"]
			// projects = ["project"]
			// user_profiles = ["user_profile"]
			// report_configuration_ids = ["report_configuration_id"]
			// accounts = ["account"]
			// fields = ["field"]
			// breakdowns = ["breakdown"]
			// action_breakdowns = ["action_breakdown"]
			// pages = ["page"]
			// repositories = ["repository"]
			// dimension_attributes = ["dimension_attribute"]
			// columns = ["column"]
			// manager_accounts = ["manager_account"]
			// profiles = ["profile"]
			// site_urls = ["site_url"]
			// api_keys = ["api_key"]
			// advertisers_id = ["advertiser_id"]
			// hosts = ["host"]
			// advertisers = ["advertiser"]
			// organizations = ["organization"]
			// account_ids = ["account_id"]

			// adobe_analytics_configurations {
			// 	sync_mode = "sync_mode"
			// 	report_suites = ["report_suite"]
			// 	elements = ["element"]
			// 	metrics = ["metric"]
			// 	calculated_metrics = ["calculated_metric"]
			// 	segments = ["segment"]
			// }
			// reports {
			// 	table = "table"
			// 	config_type = "config_type"
			// 	prebuilt_report = "prebuilt_report"
			// 	report_type = "report_type"
			// 	fields = ["field"]
			// 	dimensions = ["dimension"]
			// 	metrics = ["metric"]
			// 	segments = ["segment"]
			// 	filter = "filter"
			// }
			// custom_tables {
			// 	table_name = "table_name"
			// 	config_type = "config_type"
			// 	fields = ["field"]
			// 	breakdowns = ["breakdown"]
			// 	action_breakdowns = ["action_breakdown"]
			// 	aggregation = "aggregation"
			// 	action_report_time = "action_report_time"
			// 	click_attribution_window = "click_attribution_window"
			// 	view_attribution_window = "view_attribution_window"
			// 	prebuilt_report_name = "prebuilt_report_name"
			// }
			// project_credentials {
			// 	project = "project"
			// 	api_key = "api_key"
			// 	secret_key = "secret_key"
			// }
			// secrets_list {
			// 	key = "key"
			// 	value = "value"
			// }

			// "table":                 {Type: schema.TypeString, Computed: true},
			// "sheet_id":              {Type: schema.TypeString, Computed: true},
			// "share_url":             {Type: schema.TypeString, Computed: true},
			// "named_range":           {Type: schema.TypeString, Computed: true},
			// "client_id":             {Type: schema.TypeString, Computed: true},
			// "client_secret":         {Type: schema.TypeString, Computed: true},
			// "technical_account_id":  {Type: schema.TypeString, Computed: true},
			// "organization_id":       {Type: schema.TypeString, Computed: true},
			// "private_key":           {Type: schema.TypeString, Computed: true},
			// "sync_method":           {Type: schema.TypeString, Computed: true},
			// "sync_mode":             {Type: schema.TypeString, Computed: true},
			// "report_suites":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "elements":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "metrics":               {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "date_granularity":      {Type: schema.TypeString, Computed: true},
			// "timeframe_months":      {Type: schema.TypeString, Computed: true},
			// "source":                {Type: schema.TypeString, Computed: true},
			// "s3bucket":              {Type: schema.TypeString, Computed: true},
			// "s3role_arn":            {Type: schema.TypeString, Computed: true},
			// "abs_connection_string": {Type: schema.TypeString, Computed: true},
			// "abs_container_name":    {Type: schema.TypeString, Computed: true},
			// "folder_id":             {Type: schema.TypeString, Computed: true},
			// "ftp_host":              {Type: schema.TypeString, Computed: true},
			// "ftp_port":              {Type: schema.TypeString, Computed: true},
			// "ftp_user":              {Type: schema.TypeString, Computed: true},
			// "ftp_password":          {Type: schema.TypeString, Computed: true},
			// "is_ftps":               {Type: schema.TypeString, Computed: true},
			// "sftp_host":             {Type: schema.TypeString, Computed: true},
			// "sftp_port":             {Type: schema.TypeString, Computed: true},
			// "sftp_user":             {Type: schema.TypeString, Computed: true},
			// "sftp_password":         {Type: schema.TypeString, Computed: true},
			// "sftp_is_key_pair":      {Type: schema.TypeString, Computed: true},
			// "is_keypair":            {Type: schema.TypeString, Computed: true},
			// "advertisables":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "report_type":           {Type: schema.TypeString, Computed: true},
			// "dimensions":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "api_key":               {Type: schema.TypeString, Computed: true},
			// "external_id":           {Type: schema.TypeString, Computed: true},
			// "role_arn":              {Type: schema.TypeString, Computed: true},
			// "bucket":                {Type: schema.TypeString, Computed: true},
			// "prefix":                {Type: schema.TypeString, Computed: true},
			// "pattern":               {Type: schema.TypeString, Computed: true},
			// "file_type":             {Type: schema.TypeString, Computed: true},
			// "compression":           {Type: schema.TypeString, Computed: true},
			// "on_error":              {Type: schema.TypeString, Computed: true},
			// "append_file_option":    {Type: schema.TypeString, Computed: true},
			// "archive_pattern":       {Type: schema.TypeString, Computed: true},
			// "null_sequence":         {Type: schema.TypeString, Computed: true},
			// "delimiter":             {Type: schema.TypeString, Computed: true},
			// "escape_char":           {Type: schema.TypeString, Computed: true},
			// "skip_before":           {Type: schema.TypeString, Computed: true},
			// "skip_after":            {Type: schema.TypeString, Computed: true},
			// "project_credentials": {Type: schema.TypeList, Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"project":    {Type: schema.TypeString, Computed: true},
			// 			"api_key":    {Type: schema.TypeString, Computed: true, Sensitive: true},
			// 			"secret_key": {Type: schema.TypeString, Computed: true, Sensitive: true},
			// 		},
			// 	},
			// },
			// "auth_mode":                         {Type: schema.TypeString, Computed: true},
			// "user_name":                         {Type: schema.TypeString, Computed: true},
			// "username":                          {Type: schema.TypeString, Computed: true},
			// "password":                          {Type: schema.TypeString, Computed: true},
			// "certificate":                       {Type: schema.TypeString, Computed: true},
			// "selected_exports":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "consumer_group":                    {Type: schema.TypeString, Computed: true},
			// "servers":                           {Type: schema.TypeString, Computed: true},
			// "message_type":                      {Type: schema.TypeString, Computed: true},
			// "sync_type":                         {Type: schema.TypeString, Computed: true},
			// "security_protocol":                 {Type: schema.TypeString, Computed: true},
			// "apps":                              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "sales_accounts":                    {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "finance_accounts":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "app_sync_mode":                     {Type: schema.TypeString, Computed: true},
			// "sales_account_sync_mode":           {Type: schema.TypeString, Computed: true},
			// "finance_account_sync_mode":         {Type: schema.TypeString, Computed: true},
			// "pem_certificate":                   {Type: schema.TypeString, Computed: true},
			// "access_key_id":                     {Type: schema.TypeString, Computed: true},
			// "secret_key":                        {Type: schema.TypeString, Computed: true},
			// "home_folder":                       {Type: schema.TypeString, Computed: true},
			// "sync_data_locker":                  {Type: schema.TypeString, Computed: true},
			// "projects":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "function":                          {Type: schema.TypeString, Computed: true},
			// "region":                            {Type: schema.TypeString, Computed: true},
			// "secrets":                           {Type: schema.TypeString, Computed: true},
			// "container_name":                    {Type: schema.TypeString, Computed: true},
			// "connection_string":                 {Type: schema.TypeString, Computed: true},
			// "connection_type":                   {Type: schema.TypeString, Computed: true},
			// "function_app":                      {Type: schema.TypeString, Computed: true},
			// "function_name":                     {Type: schema.TypeString, Computed: true},
			// "function_key":                      {Type: schema.TypeString, Computed: true},
			// "public_key":                        {Type: schema.TypeString, Computed: true},
			// "merchant_id":                       {Type: schema.TypeString, Computed: true},
			// "api_url":                           {Type: schema.TypeString, Computed: true},
			// "cloud_storage_type":                {Type: schema.TypeString, Computed: true},
			// "s3external_id":                     {Type: schema.TypeString, Computed: true},
			// "s3folder":                          {Type: schema.TypeString, Computed: true},
			// "gcs_bucket":                        {Type: schema.TypeString, Computed: true},
			// "gcs_folder":                        {Type: schema.TypeString, Computed: true},
			// "user_profiles":                     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "report_configuration_ids":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "enable_all_dimension_combinations": {Type: schema.TypeString, Computed: true},
			// "instance":                          {Type: schema.TypeString, Computed: true},
			// "aws_region_code":                   {Type: schema.TypeString, Computed: true},
			// "accounts":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "fields":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "breakdowns":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "action_breakdowns":                 {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "aggregation":                       {Type: schema.TypeString, Computed: true},
			// "config_type":                       {Type: schema.TypeString, Computed: true},
			// "prebuilt_report":                   {Type: schema.TypeString, Computed: true},
			// "action_report_time":                {Type: schema.TypeString, Computed: true},
			// "click_attribution_window":          {Type: schema.TypeString, Computed: true},
			// "view_attribution_window":           {Type: schema.TypeString, Computed: true},
			// "custom_tables": {Type: schema.TypeList, Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"table_name":               {Type: schema.TypeString, Computed: true},
			// 			"config_type":              {Type: schema.TypeString, Computed: true},
			// 			"fields":                   {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"breakdowns":               {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"action_breakdowns":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"aggregation":              {Type: schema.TypeString, Computed: true},
			// 			"action_report_time":       {Type: schema.TypeString, Computed: true},
			// 			"click_attribution_window": {Type: schema.TypeString, Computed: true},
			// 			"view_attribution_window":  {Type: schema.TypeString, Computed: true},
			// 			"prebuilt_report_name":     {Type: schema.TypeString, Computed: true},
			// 		},
			// 	},
			// },
			// "pages":                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "subdomain":            {Type: schema.TypeString, Computed: true},
			// "host":                 {Type: schema.TypeString, Computed: true},
			// "port":                 {Type: schema.TypeString, Computed: true},
			// "user":                 {Type: schema.TypeString, Computed: true},
			// "is_secure":            {Type: schema.TypeString, Computed: true},
			// "repositories":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "use_webhooks":         {Type: schema.TypeString, Computed: true},
			// "dimension_attributes": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "columns":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "network_code":         {Type: schema.TypeString, Computed: true},
			// "customer_id":          {Type: schema.TypeString, Computed: true},
			// "manager_accounts":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "reports": {Type: schema.TypeList, Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"table":           {Type: schema.TypeString, Computed: true},
			// 			"config_type":     {Type: schema.TypeString, Computed: true},
			// 			"prebuilt_report": {Type: schema.TypeString, Computed: true},
			// 			"report_type":     {Type: schema.TypeString, Computed: true},
			// 			"fields":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"dimensions":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"metrics":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"segments":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"filter":          {Type: schema.TypeString, Computed: true},
			// 		},
			// 	},
			// },
			// "conversion_window_size":               {Type: schema.TypeString, Computed: true},
			// "profiles":                             {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "project_id":                           {Type: schema.TypeString, Computed: true},
			// "dataset_id":                           {Type: schema.TypeString, Computed: true},
			// "bucket_name":                          {Type: schema.TypeString, Computed: true},
			// "function_trigger":                     {Type: schema.TypeString, Computed: true},
			// "config_method":                        {Type: schema.TypeString, Computed: true},
			// "query_id":                             {Type: schema.TypeString, Computed: true},
			// "update_config_on_each_sync":           {Type: schema.TypeString, Computed: true},
			// "site_urls":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "path":                                 {Type: schema.TypeString, Computed: true},
			// "on_premise":                           {Type: schema.TypeString, Computed: true},
			// "access_token":                         {Type: schema.TypeString, Computed: true},
			// "view_through_attribution_window_size": {Type: schema.TypeString, Computed: true},
			// "post_click_attribution_window_size":   {Type: schema.TypeString, Computed: true},
			// "use_api_keys":                         {Type: schema.TypeString, Computed: true},
			// "api_keys":                             {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "endpoint":                             {Type: schema.TypeString, Computed: true},
			// "identity":                             {Type: schema.TypeString, Computed: true},
			// "api_quota":                            {Type: schema.TypeString, Computed: true},
			// "domain_name":                          {Type: schema.TypeString, Computed: true},
			// "resource_url":                         {Type: schema.TypeString, Computed: true},
			// "api_secret":                           {Type: schema.TypeString, Computed: true},
			// "hosts":                                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "tunnel_host":                          {Type: schema.TypeString, Computed: true},
			// "tunnel_port":                          {Type: schema.TypeString, Computed: true},
			// "tunnel_user":                          {Type: schema.TypeString, Computed: true},
			// "database":                             {Type: schema.TypeString, Computed: true},
			// "datasource":                           {Type: schema.TypeString, Computed: true},
			// "account":                              {Type: schema.TypeString, Computed: true},
			// "role":                                 {Type: schema.TypeString, Computed: true},
			// "email":                                {Type: schema.TypeString, Computed: true},
			// "account_id":                           {Type: schema.TypeString, Computed: true},
			// "server_url":                           {Type: schema.TypeString, Computed: true},
			// "user_key":                             {Type: schema.TypeString, Computed: true},
			// "api_version":                          {Type: schema.TypeString, Computed: true},
			// "daily_api_call_limit":                 {Type: schema.TypeString, Computed: true},
			// "time_zone":                            {Type: schema.TypeString, Computed: true},
			// "integration_key":                      {Type: schema.TypeString, Computed: true},
			// "advertisers":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "engagement_attribution_window":        {Type: schema.TypeString, Computed: true},
			// "conversion_report_time":               {Type: schema.TypeString, Computed: true},
			// "domain":                               {Type: schema.TypeString, Computed: true},
			// "update_method":                        {Type: schema.TypeString, Computed: true},
			// "replication_slot":                     {Type: schema.TypeString, Computed: true},
			// "publication_name":                     {Type: schema.TypeString, Computed: true},
			// "data_center":                          {Type: schema.TypeString, Computed: true},
			// "api_token":                            {Type: schema.TypeString, Computed: true},
			// "sub_domain":                           {Type: schema.TypeString, Computed: true},
			// "test_table_name":                      {Type: schema.TypeString, Computed: true},
			// "shop":                                 {Type: schema.TypeString, Computed: true},
			// "organizations":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "swipe_attribution_window":             {Type: schema.TypeString, Computed: true},
			// "api_access_token":                     {Type: schema.TypeString, Computed: true},
			// "account_ids":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "sid":                                  {Type: schema.TypeString, Computed: true},
			// "secret":                               {Type: schema.TypeString, Computed: true},
			// "oauth_token":                          {Type: schema.TypeString, Computed: true},
			// "oauth_token_secret":                   {Type: schema.TypeString, Computed: true},
			// "consumer_key":                         {Type: schema.TypeString, Computed: true},
			// "consumer_secret":                      {Type: schema.TypeString, Computed: true},
			// "key":                                  {Type: schema.TypeString, Computed: true},
			// "advertisers_id":                       {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// "sync_format":                          {Type: schema.TypeString, Computed: true},
			// "bucket_service":                       {Type: schema.TypeString, Computed: true},
			// "report_url":                           {Type: schema.TypeString, Computed: true},
			// "unique_id":                            {Type: schema.TypeString, Computed: true},
			// "auth_type":                            {Type: schema.TypeString, Computed: true},
			// "latest_version":                       {Type: schema.TypeString, Computed: true},
			// "authorization_method":                 {Type: schema.TypeString, Computed: true},
			// "service_version":                      {Type: schema.TypeString, Computed: true},
			// "last_synced_changes__utc_":            {Type: schema.TypeString, Computed: true},
			// "adobe_analytics_configurations": {Type: schema.TypeList, Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"sync_mode":          {Type: schema.TypeString, Computed: true},
			// 			"report_suites":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"elements":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"metrics":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"calculated_metrics": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 			"segments":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			// 		},
			// 	},
			// },
			// "is_new_package":                  {Type: schema.TypeString, Computed: true},
			// "is_multi_entity_feature_enabled": {Type: schema.TypeString, Computed: true},
			// "api_type":                        {Type: schema.TypeString, Computed: true},
			// "base_url":                        {Type: schema.TypeString, Computed: true},
			// "entity_id":                       {Type: schema.TypeString, Computed: true},
			// "soap_uri":                        {Type: schema.TypeString, Computed: true},
			// "user_id":                         {Type: schema.TypeString, Computed: true},
			// "encryption_key":                  {Type: schema.TypeString, Computed: true},
			// "always_encrypted":                {Type: schema.TypeString, Computed: true},
			// "eu_region":                       {Type: schema.TypeString, Computed: true},
			// "pat":                             {Type: schema.TypeString, Computed: true},
			// "token_key":                       {Type: schema.TypeString, Computed: true},
			// "token_secret":                    {Type: schema.TypeString, Computed: true},
			// "secrets_list": {Type: schema.TypeList, Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"key":   {Type: schema.TypeString, Computed: true},
			// 			"value": {Type: schema.TypeString, Computed: true},
			// 		},
			// 	},
			// },
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorDataSourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
