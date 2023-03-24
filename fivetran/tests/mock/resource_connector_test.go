package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	connectorEmptyMockGetHandler  *mock.Handler
	connectorEmptyMockPostHandler *mock.Handler
	connectorEmptyMockDelete      *mock.Handler

	connectorListsMockGetHandler  *mock.Handler
	connectorListsMockPostHandler *mock.Handler
	connectorListsMockDelete      *mock.Handler

	connectorMockGetHandler  *mock.Handler
	connectorMockPostHandler *mock.Handler
	connectorMockDelete      *mock.Handler

	connectorMockUpdateGetHandler   *mock.Handler
	connectorMockUpdatePostHandler  *mock.Handler
	connectorMockUpdatePatchHandler *mock.Handler
	connectorMockUpdateDelete       *mock.Handler

	connectorMockData map[string]interface{}
)

const (
	connectorWithoutConfig = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
        "sync_frequency": 5,
		"schedule_type": "auto",
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [],
            "warnings": []
        },
        "config": {}
	}
	`

	connectorUpdateResponse1 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
        "sync_frequency": 5,
		"schedule_type": "auto",
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [{
				"code":"task_code",
				"message":"task_message"
			}],
            "warnings": [{
				"code":"warning_code",
				"message":"warning_message"
			}]
        },
        "setup_tests": [{
            "title": "Validate Login",
            "status": "FAILED",
            "message": "Invalid login credentials"
        }],
        "config": {
            "user": "user",
			"password": "password"
		}
	}
	`

	connectorUpdateResponse2 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
        "paused": false,
        "pause_after_trial": false,
        "sync_frequency": 1440,
		"daily_sync_time": "3:30",
		
		"connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
		"schedule_type": "auto",

        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [{
				"code":"task_code",
				"message":"task_message"
			}],
            "warnings": [{
				"code":"warning_code",
				"message":"warning_message"
			}]
        },
        "setup_tests": [{
            "title": "Validate Login",
            "status": "FAILED",
            "message": "Invalid login credentials"
        }],
        "config": {
            "user": "user1",
			"password": "password1",
			"host": "host",
			"port": 123
		}
	}
	`

	connectorMappingResponse = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "google_sheets",
        "service_version": 1,
        "schema": "google_sheets_schema.table",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
        "sync_frequency": 5,
		"schedule_type": "auto",
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [{
				"code":"task_code",
				"message":"task_message"
			}],
            "warnings": [{
				"code":"warning_code",
				"message":"warning_message"
			}]
        },
        "setup_tests": [{
            "title": "Validate Login",
            "status": "FAILED",
            "message": "Invalid login credentials"
        }],
        "config": {
            "sheet_id": "sheet_id",
            "named_range": "range",
			"auth_type": "OAuth",

			"oauth_token":        "******",
			"oauth_token_secret": "******",
			"consumer_key":       "******",
			"client_secret":      "******",
			"private_key":        "******",
			"s3role_arn":         "******",
			"ftp_password":       "******",
			"sftp_password":      "******",
			"api_key":            "******",
			"role_arn":           "******",
			"password":           "******",
			"secret_key":         "******",
			"pem_certificate":    "******",
			"access_token":       "******",
			"api_secret":         "******",
			"api_access_token":   "******",
			"secret":             "******",
			"consumer_secret":    "******",
			"secrets":            "******",
			"api_token":          "******",
			"encryption_key":     "******",
			"pat":                "******",
			"function_trigger":   "******",
			"token_key":          "******",
			"token_secret":       "******",
			"agent_password":     "******",
			"asm_password":		  "******",

			"is_ftps":                           false,
			"sftp_is_key_pair":                  false,
			"sync_data_locker":                  false,
			"enable_all_dimension_combinations": false,
			"update_config_on_each_sync":        false,
			"on_premise":                        false,
			"is_new_package":                    false,
			"is_multi_entity_feature_enabled":   false,
			"always_encrypted":                  false,
			"use_webhooks":                      false,
			"eu_region":                         false,
			"is_keypair":                        false,
			"is_secure":                         false,
			"use_api_keys":                      false,
			"is_account_level_connector":        true,
			"use_oracle_rac": 					 false,
			"asm_option": 						 false,
			"is_single_table_mode":              true,

			"pdb_name":                         	"pdb_name",
			"agent_host":                       	"agent_host",
			"agent_user":                       	"agent_user",
			"agent_public_cert":                	"agent_public_cert",
			"agent_ora_home":						"agent_ora_home",
			"tns":									"tns",
			"asm_user":								"asm_user",
			"asm_oracle_home":						"asm_oracle_home",
			"asm_tns": 								"asm_tns",
			"sap_user": 							"sap_user",
			"connection_type":                   	"connection_type",
			"sync_mode":                         	"sync_mode",
			"date_granularity":                 	"date_granularity",
			"timeframe_months":                  	"timeframe_months",
			"report_type":                       	"report_type",
			"aggregation":                       	"aggregation",
			"config_type":                          "config_type",
			"prebuilt_report":                      "prebuilt_report",
			"action_report_time":                   "action_report_time",
			"click_attribution_window":             "click_attribution_window",
			"view_attribution_window":              "view_attribution_window",
			"view_through_attribution_window_size": "view_through_attribution_window_size",
			"post_click_attribution_window_size":   "post_click_attribution_window_size",
			"update_method":                        "update_method",
			"swipe_attribution_window":             "swipe_attribution_window",
			"api_type":                             "api_type",
			"sync_format":                          "sync_format",
			"app_sync_mode":                        "app_sync_mode",
			"sales_account_sync_mode":              "sales_account_sync_mode",
			"finance_account_sync_mode":            "finance_account_sync_mode",
			"source":                               "source",
			"file_type":                            "file_type",
			"compression":                          "compression",
			"on_error":                             "on_error",
			"append_file_option":                   "append_file_option",
			"engagement_attribution_window":        "engagement_attribution_window",
			"conversion_report_time":               "conversion_report_time",

			"conversion_window_size":               0,
			"skip_before":                          0,
			"skip_after":                           0,
			"ftp_port":                             0,
			"sftp_port":             				0,
			"port":                 				0,
			"tunnel_port":                          0,
			"api_quota":                            0,
			"daily_api_call_limit":                 0,
			"agent_port":                           0,

			"public_key": 			"public_key",
			"external_id": 			"external_id",
			"group_name":           "group_name",
			"client_id":             "client_id",
			"technical_account_id":  "technical_account_id",
			"organization_id":       "organization_id",
			"s3bucket":              "s3bucket",
			"abs_connection_string": "abs_connection_string",
			"abs_container_name":    "abs_container_name",
			"folder_id":             "folder_id",
			"ftp_host":              "ftp_host",
			"ftp_user":              "ftp_user",
			"sftp_host":             "sftp_host",
			"sync_method":           "sync_method",
			
			"sftp_user":             "sftp_user",
			"bucket":                "bucket",
			"prefix":                "prefix",
			"pattern":               "pattern",
			"archive_pattern":       "archive_pattern",
			"null_sequence":         "null_sequence",
			"delimiter":             "delimiter",
			"escape_char":           "escape_char",
			"auth_mode":             "auth_mode",
			"certificate":           "certificate",
			"consumer_group":        "consumer_group",
			"servers":               "servers",
			"message_type":          "message_type",
			"sync_type":             "sync_type",
			"security_protocol":     "security_protocol",
			"access_key_id":         "access_key_id",
			"home_folder":           "home_folder",
			"function":              "function",
			"region":                "region",
			"container_name":        "container_name",
			"connection_string":     "connection_string",
			"function_app":          "function_app",
			"function_name":         "function_name",
			"function_key":          "function_key",
			"merchant_id":           "merchant_id",
			"api_url":               "api_url",
			"cloud_storage_type":    "cloud_storage_type",
			"s3external_id":         "s3external_id",
			"s3folder":              "s3folder",
			"gcs_bucket":            "gcs_bucket",
			"gcs_folder":            "gcs_folder",
			"instance":              "instance",
			"aws_region_code":       "aws_region_code",
			"subdomain":             "subdomain",
			"host":                  "host",

			"user":                 "user",
			"network_code":         "network_code",
			"customer_id":          "customer_id",
			"project_id":           "project_id",
			"dataset_id":           "dataset_id",
			"bucket_name":          "bucket_name",
			"config_method":        "config_method",
			"query_id":             "query_id",
			"path":                 "path",
			"endpoint":             "endpoint",
			"identity":             "identity",
			
			"domain_name":          "domain_name",
			"resource_url":         "resource_url",
			"tunnel_host":          "tunnel_host",
			"tunnel_user":          "tunnel_user",
			"database":             "database",
			"datasource":           "datasource",
			"account":              "account",
			"role":                 "role",
			"email":                "email",
			"account_id":           "account_id",
			"server_url":           "server_url",
			"user_key":             "user_key",
			"api_version":          "api_version",
			"time_zone":            "time_zone",
			"integration_key":      "integration_key",
			"domain":               "domain",
			"replication_slot":     "replication_slot",
			"publication_name":     "publication_name",
			"data_center":          "data_center",
			"sub_domain":           "sub_domain",
			"test_table_name":      "test_table_name",
			"shop":                 "shop",
			"sid":                  "sid",
			"key":                  "key",
			"bucket_service":       "bucket_service",
			"user_name":            "user_name",
			"username":             "username",
			"report_url":           "report_url",
			"unique_id":            "unique_id",
			"base_url":             "base_url",
			"entity_id":            "entity_id",
			"soap_uri":             "soap_uri",
			"user_id":              "user_id",
			"share_url":            "share_url",
			"organization":         "organization",
			"access_key":           "access_key",
			"domain_host_name":     "domain_host_name",
			"client_name":          "client_name",
			"domain_type":          "domain_type",
			"connection_method":    "connection_method",

			"report_suites":            ["report_suite"],
			"elements":                 ["element"],
			"metrics":                  ["metric"],
			"advertisables":     		["advertisable"],
			"dimensions": 				["dimension"],
			"selected_exports": 		["selected_export"],
			"apps": 					["app"],
			"sales_accounts": 			["sales_account"],
			"finance_accounts": 		["finance_account"],
			"projects": 				["project"],
			"user_profiles": 			["user_profile"],
			"report_configuration_ids": ["report_configuration_id"],
			"accounts": 				["account"],
			"fields": 					["field"],
			"breakdowns": 				["breakdown"],
			"action_breakdowns": 		["action_breakdown"],
			"pages": 					["page"],
			"repositories": 			["repository"],
			"dimension_attributes": 	["dimension_attribute"],
			"columns": 					["column"],
			"manager_accounts": 		["manager_account"],
			"profiles": 				["profile"],
			"site_urls": 				["site_url"],
			"api_keys": 				["******"],
			"advertisers_id": 			["advertiser_id"],
			"hosts": 					["host"],
			"advertisers": 				["advertiser"],
			"organizations": 			["organization"],
			"account_ids": 				["account_id"],
			"packed_mode_tables":       ["packed_mode_table"],

			"adobe_analytics_configurations": [{
				"sync_mode": 			"sync_mode",
				"report_suites": 		["report_suite"],
				"elements": 			["element"],
				"metrics": 				["metric"],
				"calculated_metrics": 	["calculated_metric"],
				"segments": 			["segment"]
			}],
			"reports": [{
				"table": 			"table",
				"config_type": 		"config_type",
				"prebuilt_report": 	"prebuilt_report",
				"report_type": 		"report_type",
				"fields": 			["field"],
				"dimensions": 		["dimension"],
				"metrics": 			["metric"],
				"segments": 		["segment"],
				"filter": 			"filter"
			}],
			"custom_tables": [{
				"table_name": 				"table_name",
				"config_type": 				"config_type",
				"fields": 					["field"],
				"breakdowns": 				["breakdown"],
				"action_breakdowns": 		["action_breakdown"],
				"aggregation": 				"aggregation",
				"action_report_time": 		"action_report_time",
				"click_attribution_window": "click_attribution_window",
				"view_attribution_window": 	"view_attribution_window",
				"prebuilt_report_name": 	"prebuilt_report_name"
			}],
			"project_credentials": [{
				"project": 		"project",
				"api_key": 		"api_key",
				"secret_key": 	"******"
			}],
			"secrets_list": [{
				"key":   "key",
				"value": "******"
			}]
        }
	}
	`

	connectorConfigMappingTfConfig = `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service = "google_sheets"

		destination_schema {
			name = "google_sheets_schema"
			table = "table"
		}

		sync_frequency = 5
		paused = true
		pause_after_trial = true
		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		config {
			sheet_id = "sheet_id"
			named_range = "range"
			auth_type = "OAuth"

			oauth_token = "oauth_token" 
			oauth_token_secret = "oauth_token_secret"
			consumer_key = "consumer_key"
			client_secret = "client_secret"
			private_key = "private_key"
			s3role_arn = "s3role_arn"
			ftp_password = "ftp_password"
			sftp_password = "sftp_password"
			api_key = "api_key"
			role_arn = "role_arn"
			password = "password"
			secret_key = "secret_key"
			pem_certificate = "pem_certificate"
			access_token = "access_token"
			api_secret = "api_secret"
			api_access_token = "api_access_token"
			secret = "secret"
			consumer_secret = "consumer_secret"
			secrets = "secrets"
			api_token = "api_token"
			encryption_key = "encryption_key"
			pat = "pat"
			function_trigger = "function_trigger"
			token_key = "token_key"
			token_secret = "token_secret"

			sync_method = "sync_method"

			is_ftps = "false"
			sftp_is_key_pair = "false"
			sync_data_locker = "false"
			enable_all_dimension_combinations = "false"
			update_config_on_each_sync = "false"
			on_premise = "false"
			is_new_package = "false"
			is_multi_entity_feature_enabled = "false"
			always_encrypted = "false"
			is_secure = "false"
			use_api_keys = "false"
			use_webhooks = "false"
			eu_region = "false"
			is_keypair = "false"
			is_account_level_connector = "true"

			conversion_window_size = "0"
			skip_before = "0"
			skip_after = "0"
			ftp_port = "0"
			sftp_port = "0"
			port = "0"
			tunnel_port = "0"
			api_quota = "0"
			daily_api_call_limit = "0"
			agent_port = "0"

			pdb_name = "pdb_name"
			agent_host = "agent_host"
			agent_user = "agent_user"
			agent_password = "agent_password"
			agent_public_cert = "agent_public_cert"
			agent_ora_home = "agent_ora_home"
			tns = "tns"
			use_oracle_rac = "false"
			is_single_table_mode = "true"
			asm_option = "false"
		    asm_user = "asm_user"
			asm_password = "asm_password"
			asm_oracle_home = "asm_oracle_home"
			asm_tns = "asm_tns"
			sap_user = "sap_user"
			group_name = "group_name"

			connection_type = "connection_type"
			sync_mode = "sync_mode"
			date_granularity = "date_granularity"
			timeframe_months = "timeframe_months"
			report_type = "report_type"
			aggregation = "aggregation"
			config_type = "config_type"
			prebuilt_report = "prebuilt_report"
			action_report_time = "action_report_time"
			click_attribution_window = "click_attribution_window"
			view_attribution_window = "view_attribution_window"
			view_through_attribution_window_size = "view_through_attribution_window_size"
			post_click_attribution_window_size = "post_click_attribution_window_size"
			update_method = "update_method"
			swipe_attribution_window = "swipe_attribution_window"
			api_type = "api_type"
			sync_format = "sync_format"
			app_sync_mode = "app_sync_mode"
			sales_account_sync_mode = "sales_account_sync_mode"
			finance_account_sync_mode = "finance_account_sync_mode"
			source = "source"
			file_type = "file_type"
			compression = "compression"
			on_error = "on_error"
			append_file_option = "append_file_option"
			engagement_attribution_window = "engagement_attribution_window"
			conversion_report_time = "conversion_report_time"

			external_id = "external_id"
			public_key = "public_key"

			client_id = "client_id"
			technical_account_id = "technical_account_id"
			organization_id = "organization_id"
			s3bucket = "s3bucket"
			abs_connection_string = "abs_connection_string"
			abs_container_name = "abs_container_name"
			folder_id = "folder_id"
			ftp_host = "ftp_host"
			ftp_user = "ftp_user"
			sftp_host = "sftp_host"
			sftp_user = "sftp_user"
			bucket = "bucket"
			prefix = "prefix"
			pattern = "pattern"
			archive_pattern = "archive_pattern"
			null_sequence = "null_sequence"
			delimiter = "delimiter"
			escape_char = "escape_char"
			auth_mode = "auth_mode"
			certificate = "certificate"
			consumer_group = "consumer_group"
			servers = "servers"
			message_type = "message_type"
			sync_type = "sync_type"
			security_protocol = "security_protocol"
			access_key_id = "access_key_id"
			home_folder = "home_folder"
			function = "function"
			region = "region"
			container_name = "container_name"
			connection_string = "connection_string"
			function_app = "function_app"
			function_name = "function_name"
			function_key = "function_key"
			merchant_id = "merchant_id"
			api_url = "api_url"
			cloud_storage_type = "cloud_storage_type"
			s3external_id = "s3external_id"
			s3folder = "s3folder"
			gcs_bucket = "gcs_bucket"
			gcs_folder = "gcs_folder"
			instance = "instance"
			aws_region_code = "aws_region_code"
			host = "host"

			user = "user"
			network_code = "network_code"
			customer_id = "customer_id"
			project_id = "project_id"
			dataset_id = "dataset_id"
			bucket_name = "bucket_name"
			config_method = "config_method"
			query_id = "query_id"
			path = "path"
			endpoint = "endpoint"
			identity = "identity"
			
			domain_name = "domain_name"
			resource_url = "resource_url"
			tunnel_host = "tunnel_host"
			tunnel_user = "tunnel_user"
			database = "database"
			datasource = "datasource"
			account = "account"
			role = "role"
			email = "email"
			account_id = "account_id"
			server_url = "server_url"
			user_key = "user_key"
			api_version = "api_version"
			time_zone = "time_zone"
			integration_key = "integration_key"
			domain = "domain"
			replication_slot = "replication_slot"
			publication_name = "publication_name"
			data_center = "data_center"
			sub_domain = "sub_domain"
			subdomain = "subdomain"
			test_table_name = "test_table_name"
			shop = "shop"
			sid = "sid"
			key = "key"
			bucket_service = "bucket_service"
			user_name = "user_name"
			username = "username"
			report_url = "report_url"
			unique_id = "unique_id"
			base_url = "base_url"
			entity_id = "entity_id"
			soap_uri = "soap_uri"
			user_id = "user_id"
			share_url = "share_url"
			organization = "organization"
			access_key = "access_key"
			domain_host_name = "domain_host_name"
			client_name = "client_name"
			domain_type = "domain_type"
			connection_method = "connection_method"

			report_suites = ["report_suite"]
			elements = ["element"]
			metrics = ["metric"]
			advertisables = ["advertisable"]
			dimensions = ["dimension"]
			selected_exports = ["selected_export"]
			apps = ["app"]
			sales_accounts = ["sales_account"]
			finance_accounts = ["finance_account"]
			projects = ["project"]
			user_profiles = ["user_profile"]
			report_configuration_ids = ["report_configuration_id"]
			accounts = ["account"]
			fields = ["field"]
			breakdowns = ["breakdown"]
			action_breakdowns = ["action_breakdown"]
			pages = ["page"]
			repositories = ["repository"]
			dimension_attributes = ["dimension_attribute"]
			columns = ["column"]
			manager_accounts = ["manager_account"]
			profiles = ["profile"]
			site_urls = ["site_url"]
			api_keys = ["api_key"]
			advertisers_id = ["advertiser_id"]
			hosts = ["host"]
			advertisers = ["advertiser"]
			organizations = ["organization"]
			account_ids = ["account_id"]
			packed_mode_tables = ["packed_mode_table"]


			adobe_analytics_configurations {
				sync_mode = "sync_mode"
				report_suites = ["report_suite"]
				elements = ["element"]
				metrics = ["metric"]
				calculated_metrics = ["calculated_metric"]
				segments = ["segment"]
			}
			reports {
				table = "table"
				config_type = "config_type"
				prebuilt_report = "prebuilt_report"
				report_type = "report_type"
				fields = ["field"]
				dimensions = ["dimension"]
				metrics = ["metric"]
				segments = ["segment"]
				filter = "filter"
			}
			custom_tables {
				table_name = "table_name"
				config_type = "config_type"
				fields = ["field"]
				breakdowns = ["breakdown"]
				action_breakdowns = ["action_breakdown"]
				aggregation = "aggregation"
				action_report_time = "action_report_time"
				click_attribution_window = "click_attribution_window"
				view_attribution_window = "view_attribution_window"
				prebuilt_report_name = "prebuilt_report_name"
			}
			project_credentials {
				project = "project"
				api_key = "api_key"
				secret_key = "secret_key"
			}
			secrets_list {
				key = "key"
				value = "value"
			}
		}

		auth {
			refresh_token = "refresh_token"
			access_token = "access_token"
			realm_id = "realm_id"
			client_access {
				client_id = "client_id"
				client_secret = "client_secret"
				user_agent = "user_agent"
				developer_token = "developer_token"
			}
		}
	}
	`

	connectorConfigListsMappingResponse = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "google_sheets",
        "service_version": 1,
        "schema": "google_sheets_schema.table",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
        "sync_frequency": 5,
		"schedule_type": "auto",
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [{
				"code":"task_code",
				"message":"task_message"
			}],
            "warnings": [{
				"code":"warning_code",
				"message":"warning_message"
			}]
        },
        "setup_tests": [{
            "title": "Validate Login",
            "status": "FAILED",
            "message": "Invalid login credentials"
        }],
        "config": {
			"packed_mode_tables":["packed_mode_table_3", "packed_mode_table_2", "packed_mode_table_1"],
			"report_suites": ["value_2", "value_1"],
			"elements": ["value_2", "value_1"],
			"metrics": ["value_2", "value_1"],
			"advertisables": ["value_2", "value_1"],
			"dimensions": ["value_2", "value_1"],
			"selected_exports": ["value_2", "value_1"],
			"apps": ["value_2", "value_1"],
			"sales_accounts": ["value_2", "value_1"],
			"finance_accounts": ["value_2", "value_1"],
			"projects": ["value_2", "value_1"],
			"user_profiles": ["value_2", "value_1"],
			"report_configuration_ids": ["value_2", "value_1"],
			"accounts": ["value_2", "value_1"],
			"fields": ["value_2", "value_1"],
			"breakdowns": ["value_2", "value_1"],
			"action_breakdowns": ["value_2", "value_1"],
			"pages": ["value_2", "value_1"],
			"repositories": ["value_2", "value_1"],
			"dimension_attributes": ["value_2", "value_1"],
			"columns": ["value_2", "value_1"],
			"manager_accounts": ["value_2", "value_1"],
			"profiles": ["value_2", "value_1"],
			"site_urls": ["value_2", "value_1"],
			"api_keys": ["value_2", "value_1"],
			"advertisers_id": ["value_2", "value_1"],
			"hosts": ["value_2", "value_1"],
			"advertisers": ["value_2", "value_1"],
			"organizations": ["value_2", "value_1"],
			"account_ids": ["value_2", "value_1"]
		}
	}
	`

	connectorConfigListsMappingTfConfig = `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service = "google_sheets"

		destination_schema {
			name = "google_sheets_schema"
			table = "table"
		}

		sync_frequency = 5
		paused = true
		pause_after_trial = true
		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		config {
			packed_mode_tables = ["packed_mode_table_1", "packed_mode_table_2", "packed_mode_table_3"]
			report_suites = ["value_1", "value_2"]
			elements = ["value_1", "value_2"]
			metrics = ["value_1", "value_2"]
			advertisables = ["value_1", "value_2"]
			dimensions = ["value_1", "value_2"]
			selected_exports = ["value_1", "value_2"]
			apps = ["value_1", "value_2"]
			sales_accounts = ["value_1", "value_2"]
			finance_accounts = ["value_1", "value_2"]
			projects = ["value_1", "value_2"]
			user_profiles = ["value_1", "value_2"]
			report_configuration_ids = ["value_1", "value_2"]
			accounts = ["value_1", "value_2"]
			fields = ["value_1", "value_2"]
			breakdowns = ["value_1", "value_2"]
			action_breakdowns = ["value_1", "value_2"]
			pages = ["value_1", "value_2"]
			repositories = ["value_1", "value_2"]
			dimension_attributes = ["value_1", "value_2"]
			columns = ["value_1", "value_2"]
			manager_accounts = ["value_1", "value_2"]
			profiles = ["value_1", "value_2"]
			site_urls = ["value_1", "value_2"]
			api_keys = ["value_1", "value_2"]
			advertisers_id = ["value_1", "value_2"]
			hosts = ["value_1", "value_2"]
			advertisers = ["value_1", "value_2"]
			organizations = ["value_1", "value_2"]
			account_ids = ["value_1", "value_2"]
		}
	}
	`
)

func setupMockClientConnectorResourceListMappingConfig(t *testing.T) {
	mockClient.Reset()

	connectorListsMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorListsMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = createMapFromJsonString(t, connectorConfigListsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorListsMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	connectorMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			assertKeyExists(t, body, "config")

			config := body["config"].(map[string]interface{})

			assertKeyExistsAndHasValue(t, config, "sheet_id", "sheet_id")
			assertKeyExistsAndHasValue(t, config, "named_range", "range")
			assertKeyExistsAndHasValue(t, config, "auth_type", "OAuth")

			assertKeyExistsAndHasValue(t, config, "oauth_token", "oauth_token")
			assertKeyExistsAndHasValue(t, config, "oauth_token_secret", "oauth_token_secret")
			assertKeyExistsAndHasValue(t, config, "consumer_key", "consumer_key")
			assertKeyExistsAndHasValue(t, config, "client_secret", "client_secret")
			assertKeyExistsAndHasValue(t, config, "private_key", "private_key")
			assertKeyExistsAndHasValue(t, config, "s3role_arn", "s3role_arn")
			assertKeyExistsAndHasValue(t, config, "ftp_password", "ftp_password")
			assertKeyExistsAndHasValue(t, config, "sftp_password", "sftp_password")
			assertKeyExistsAndHasValue(t, config, "api_key", "api_key")
			assertKeyExistsAndHasValue(t, config, "role_arn", "role_arn")
			assertKeyExistsAndHasValue(t, config, "password", "password")
			assertKeyExistsAndHasValue(t, config, "secret_key", "secret_key")
			assertKeyExistsAndHasValue(t, config, "pem_certificate", "pem_certificate")
			assertKeyExistsAndHasValue(t, config, "access_token", "access_token")
			assertKeyExistsAndHasValue(t, config, "api_secret", "api_secret")
			assertKeyExistsAndHasValue(t, config, "api_access_token", "api_access_token")
			assertKeyExistsAndHasValue(t, config, "secret", "secret")
			assertKeyExistsAndHasValue(t, config, "consumer_secret", "consumer_secret")
			assertKeyExistsAndHasValue(t, config, "secrets", "secrets")
			assertKeyExistsAndHasValue(t, config, "api_token", "api_token")
			assertKeyExistsAndHasValue(t, config, "encryption_key", "encryption_key")
			assertKeyExistsAndHasValue(t, config, "pat", "pat")
			assertKeyExistsAndHasValue(t, config, "function_trigger", "function_trigger")
			assertKeyExistsAndHasValue(t, config, "token_key", "token_key")
			assertKeyExistsAndHasValue(t, config, "token_secret", "token_secret")
			assertKeyExistsAndHasValue(t, config, "asm_password", "asm_password")
			assertKeyExistsAndHasValue(t, config, "agent_password", "agent_password")

			assertKeyExistsAndHasValue(t, config, "is_ftps", false)
			assertKeyExistsAndHasValue(t, config, "sftp_is_key_pair", false)
			assertKeyExistsAndHasValue(t, config, "sync_data_locker", false)
			assertKeyExistsAndHasValue(t, config, "enable_all_dimension_combinations", false)
			assertKeyExistsAndHasValue(t, config, "update_config_on_each_sync", false)
			assertKeyExistsAndHasValue(t, config, "on_premise", false)
			assertKeyExistsAndHasValue(t, config, "is_new_package", false)
			assertKeyExistsAndHasValue(t, config, "is_multi_entity_feature_enabled", false)
			assertKeyExistsAndHasValue(t, config, "always_encrypted", false)
			assertKeyExistsAndHasValue(t, config, "use_webhooks", false)
			assertKeyExistsAndHasValue(t, config, "eu_region", false)
			assertKeyExistsAndHasValue(t, config, "is_secure", false)
			assertKeyExistsAndHasValue(t, config, "use_api_keys", false)
			assertKeyExistsAndHasValue(t, config, "is_keypair", false)
			assertKeyExistsAndHasValue(t, config, "is_account_level_connector", true)
			assertKeyExistsAndHasValue(t, config, "use_oracle_rac", false)
			assertKeyExistsAndHasValue(t, config, "asm_option", false)
			assertKeyExistsAndHasValue(t, config, "is_single_table_mode", true)

			assertKeyExistsAndHasValue(t, config, "connection_type", "connection_type")
			assertKeyExistsAndHasValue(t, config, "sync_mode", "sync_mode")
			assertKeyExistsAndHasValue(t, config, "date_granularity", "date_granularity")
			assertKeyExistsAndHasValue(t, config, "timeframe_months", "timeframe_months")
			assertKeyExistsAndHasValue(t, config, "report_type", "report_type")
			assertKeyExistsAndHasValue(t, config, "aggregation", "aggregation")
			assertKeyExistsAndHasValue(t, config, "config_type", "config_type")
			assertKeyExistsAndHasValue(t, config, "prebuilt_report", "prebuilt_report")
			assertKeyExistsAndHasValue(t, config, "action_report_time", "action_report_time")
			assertKeyExistsAndHasValue(t, config, "click_attribution_window", "click_attribution_window")
			assertKeyExistsAndHasValue(t, config, "view_attribution_window", "view_attribution_window")
			assertKeyExistsAndHasValue(t, config, "view_through_attribution_window_size", "view_through_attribution_window_size")
			assertKeyExistsAndHasValue(t, config, "post_click_attribution_window_size", "post_click_attribution_window_size")
			assertKeyExistsAndHasValue(t, config, "update_method", "update_method")
			assertKeyExistsAndHasValue(t, config, "swipe_attribution_window", "swipe_attribution_window")
			assertKeyExistsAndHasValue(t, config, "api_type", "api_type")
			assertKeyExistsAndHasValue(t, config, "sync_format", "sync_format")
			assertKeyExistsAndHasValue(t, config, "app_sync_mode", "app_sync_mode")
			assertKeyExistsAndHasValue(t, config, "sales_account_sync_mode", "sales_account_sync_mode")
			assertKeyExistsAndHasValue(t, config, "finance_account_sync_mode", "finance_account_sync_mode")
			assertKeyExistsAndHasValue(t, config, "source", "source")
			assertKeyExistsAndHasValue(t, config, "file_type", "file_type")
			assertKeyExistsAndHasValue(t, config, "compression", "compression")
			assertKeyExistsAndHasValue(t, config, "on_error", "on_error")
			assertKeyExistsAndHasValue(t, config, "append_file_option", "append_file_option")
			assertKeyExistsAndHasValue(t, config, "engagement_attribution_window", "engagement_attribution_window")
			assertKeyExistsAndHasValue(t, config, "conversion_report_time", "conversion_report_time")

			// all numbers in json are float64
			assertKeyExistsAndHasValue(t, config, "conversion_window_size", float64(0))
			assertKeyExistsAndHasValue(t, config, "skip_before", float64(0))
			assertKeyExistsAndHasValue(t, config, "skip_after", float64(0))
			assertKeyExistsAndHasValue(t, config, "ftp_port", float64(0))
			assertKeyExistsAndHasValue(t, config, "sftp_port", float64(0))
			assertKeyExistsAndHasValue(t, config, "port", float64(0))
			assertKeyExistsAndHasValue(t, config, "agent_port", float64(0))
			assertKeyExistsAndHasValue(t, config, "tunnel_port", float64(0))
			assertKeyExistsAndHasValue(t, config, "api_quota", float64(0))
			assertKeyExistsAndHasValue(t, config, "daily_api_call_limit", float64(0))

			assertKeyExistsAndHasValue(t, config, "group_name", "group_name")
			assertKeyExistsAndHasValue(t, config, "pdb_name", "pdb_name")
			assertKeyExistsAndHasValue(t, config, "agent_host", "agent_host")
			assertKeyExistsAndHasValue(t, config, "agent_user", "agent_user")
			assertKeyExistsAndHasValue(t, config, "agent_public_cert", "agent_public_cert")
			assertKeyExistsAndHasValue(t, config, "agent_ora_home", "agent_ora_home")
			assertKeyExistsAndHasValue(t, config, "tns", "tns")
			assertKeyExistsAndHasValue(t, config, "asm_user", "asm_user")
			assertKeyExistsAndHasValue(t, config, "asm_oracle_home", "asm_oracle_home")
			assertKeyExistsAndHasValue(t, config, "asm_tns", "asm_tns")
			assertKeyExistsAndHasValue(t, config, "sap_user", "sap_user")

			assertKeyExistsAndHasValue(t, config, "public_key", "public_key")
			assertKeyExistsAndHasValue(t, config, "external_id", "external_id")
			assertKeyExistsAndHasValue(t, config, "client_id", "client_id")
			assertKeyExistsAndHasValue(t, config, "technical_account_id", "technical_account_id")
			assertKeyExistsAndHasValue(t, config, "organization_id", "organization_id")
			assertKeyExistsAndHasValue(t, config, "s3bucket", "s3bucket")
			assertKeyExistsAndHasValue(t, config, "abs_connection_string", "abs_connection_string")
			assertKeyExistsAndHasValue(t, config, "abs_container_name", "abs_container_name")
			assertKeyExistsAndHasValue(t, config, "folder_id", "folder_id")
			assertKeyExistsAndHasValue(t, config, "ftp_host", "ftp_host")
			assertKeyExistsAndHasValue(t, config, "ftp_user", "ftp_user")
			assertKeyExistsAndHasValue(t, config, "sftp_host", "sftp_host")
			assertKeyExistsAndHasValue(t, config, "sftp_user", "sftp_user")
			assertKeyExistsAndHasValue(t, config, "bucket", "bucket")
			assertKeyExistsAndHasValue(t, config, "prefix", "prefix")
			assertKeyExistsAndHasValue(t, config, "pattern", "pattern")
			assertKeyExistsAndHasValue(t, config, "archive_pattern", "archive_pattern")
			assertKeyExistsAndHasValue(t, config, "null_sequence", "null_sequence")
			assertKeyExistsAndHasValue(t, config, "delimiter", "delimiter")
			assertKeyExistsAndHasValue(t, config, "escape_char", "escape_char")
			assertKeyExistsAndHasValue(t, config, "auth_mode", "auth_mode")
			assertKeyExistsAndHasValue(t, config, "certificate", "certificate")
			assertKeyExistsAndHasValue(t, config, "consumer_group", "consumer_group")
			assertKeyExistsAndHasValue(t, config, "servers", "servers")
			assertKeyExistsAndHasValue(t, config, "message_type", "message_type")
			assertKeyExistsAndHasValue(t, config, "sync_type", "sync_type")
			assertKeyExistsAndHasValue(t, config, "security_protocol", "security_protocol")
			assertKeyExistsAndHasValue(t, config, "access_key_id", "access_key_id")
			assertKeyExistsAndHasValue(t, config, "home_folder", "home_folder")
			assertKeyExistsAndHasValue(t, config, "function", "function")
			assertKeyExistsAndHasValue(t, config, "region", "region")
			assertKeyExistsAndHasValue(t, config, "container_name", "container_name")
			assertKeyExistsAndHasValue(t, config, "connection_string", "connection_string")
			assertKeyExistsAndHasValue(t, config, "function_app", "function_app")
			assertKeyExistsAndHasValue(t, config, "function_name", "function_name")
			assertKeyExistsAndHasValue(t, config, "function_key", "function_key")
			assertKeyExistsAndHasValue(t, config, "merchant_id", "merchant_id")
			assertKeyExistsAndHasValue(t, config, "api_url", "api_url")
			assertKeyExistsAndHasValue(t, config, "cloud_storage_type", "cloud_storage_type")
			assertKeyExistsAndHasValue(t, config, "s3external_id", "s3external_id")
			assertKeyExistsAndHasValue(t, config, "s3folder", "s3folder")
			assertKeyExistsAndHasValue(t, config, "gcs_bucket", "gcs_bucket")
			assertKeyExistsAndHasValue(t, config, "gcs_folder", "gcs_folder")
			assertKeyExistsAndHasValue(t, config, "instance", "instance")
			assertKeyExistsAndHasValue(t, config, "aws_region_code", "aws_region_code")
			assertKeyExistsAndHasValue(t, config, "sub_domain", "sub_domain")
			assertKeyExistsAndHasValue(t, config, "subdomain", "subdomain")
			assertKeyExistsAndHasValue(t, config, "host", "host")
			assertKeyExistsAndHasValue(t, config, "user", "user")
			assertKeyExistsAndHasValue(t, config, "network_code", "network_code")
			assertKeyExistsAndHasValue(t, config, "customer_id", "customer_id")
			assertKeyExistsAndHasValue(t, config, "project_id", "project_id")
			assertKeyExistsAndHasValue(t, config, "dataset_id", "dataset_id")
			assertKeyExistsAndHasValue(t, config, "bucket_name", "bucket_name")
			assertKeyExistsAndHasValue(t, config, "config_method", "config_method")
			assertKeyExistsAndHasValue(t, config, "query_id", "query_id")
			assertKeyExistsAndHasValue(t, config, "path", "path")
			assertKeyExistsAndHasValue(t, config, "endpoint", "endpoint")
			assertKeyExistsAndHasValue(t, config, "identity", "identity")
			assertKeyExistsAndHasValue(t, config, "domain_name", "domain_name")
			assertKeyExistsAndHasValue(t, config, "resource_url", "resource_url")
			assertKeyExistsAndHasValue(t, config, "tunnel_host", "tunnel_host")
			assertKeyExistsAndHasValue(t, config, "tunnel_user", "tunnel_user")
			assertKeyExistsAndHasValue(t, config, "database", "database")
			assertKeyExistsAndHasValue(t, config, "datasource", "datasource")
			assertKeyExistsAndHasValue(t, config, "account", "account")
			assertKeyExistsAndHasValue(t, config, "role", "role")
			assertKeyExistsAndHasValue(t, config, "email", "email")
			assertKeyExistsAndHasValue(t, config, "account_id", "account_id")
			assertKeyExistsAndHasValue(t, config, "server_url", "server_url")
			assertKeyExistsAndHasValue(t, config, "user_key", "user_key")
			assertKeyExistsAndHasValue(t, config, "api_version", "api_version")
			assertKeyExistsAndHasValue(t, config, "time_zone", "time_zone")
			assertKeyExistsAndHasValue(t, config, "integration_key", "integration_key")
			assertKeyExistsAndHasValue(t, config, "domain", "domain")
			assertKeyExistsAndHasValue(t, config, "replication_slot", "replication_slot")
			assertKeyExistsAndHasValue(t, config, "publication_name", "publication_name")
			assertKeyExistsAndHasValue(t, config, "data_center", "data_center")
			assertKeyExistsAndHasValue(t, config, "sub_domain", "sub_domain")
			assertKeyExistsAndHasValue(t, config, "test_table_name", "test_table_name")
			assertKeyExistsAndHasValue(t, config, "shop", "shop")
			assertKeyExistsAndHasValue(t, config, "sid", "sid")
			assertKeyExistsAndHasValue(t, config, "key", "key")
			assertKeyExistsAndHasValue(t, config, "bucket_service", "bucket_service")
			assertKeyExistsAndHasValue(t, config, "user_name", "user_name")
			assertKeyExistsAndHasValue(t, config, "username", "username")
			assertKeyExistsAndHasValue(t, config, "report_url", "report_url")
			assertKeyExistsAndHasValue(t, config, "unique_id", "unique_id")
			assertKeyExistsAndHasValue(t, config, "base_url", "base_url")
			assertKeyExistsAndHasValue(t, config, "entity_id", "entity_id")
			assertKeyExistsAndHasValue(t, config, "soap_uri", "soap_uri")
			assertKeyExistsAndHasValue(t, config, "user_id", "user_id")
			assertKeyExistsAndHasValue(t, config, "share_url", "share_url")
			assertKeyExistsAndHasValue(t, config, "organization", "organization")
			assertKeyExistsAndHasValue(t, config, "access_key", "access_key")
			assertKeyExistsAndHasValue(t, config, "domain_host_name", "domain_host_name")
			assertKeyExistsAndHasValue(t, config, "client_name", "client_name")
			assertKeyExistsAndHasValue(t, config, "domain_type", "domain_type")
			assertKeyExistsAndHasValue(t, config, "connection_method", "connection_method")
			assertKeyExists(t, config, "report_suites")
			assertArrayItems(t, config["report_suites"].([]interface{}), append(make([]interface{}, 0), "report_suite"))

			assertKeyExists(t, config, "elements")
			assertArrayItems(t, config["elements"].([]interface{}), append(make([]interface{}, 0), "element"))

			assertKeyExists(t, config, "metrics")
			assertArrayItems(t, config["metrics"].([]interface{}), append(make([]interface{}, 0), "metric"))

			assertKeyExists(t, config, "advertisables")
			assertArrayItems(t, config["advertisables"].([]interface{}), append(make([]interface{}, 0), "advertisable"))

			assertKeyExists(t, config, "dimensions")
			assertArrayItems(t, config["dimensions"].([]interface{}), append(make([]interface{}, 0), "dimension"))

			assertKeyExists(t, config, "selected_exports")
			assertArrayItems(t, config["selected_exports"].([]interface{}), append(make([]interface{}, 0), "selected_export"))

			assertKeyExists(t, config, "apps")
			assertArrayItems(t, config["apps"].([]interface{}), append(make([]interface{}, 0), "app"))

			assertKeyExists(t, config, "sales_accounts")
			assertArrayItems(t, config["sales_accounts"].([]interface{}), append(make([]interface{}, 0), "sales_account"))

			assertKeyExists(t, config, "finance_accounts")
			assertArrayItems(t, config["finance_accounts"].([]interface{}), append(make([]interface{}, 0), "finance_account"))

			assertKeyExists(t, config, "projects")
			assertArrayItems(t, config["projects"].([]interface{}), append(make([]interface{}, 0), "project"))

			assertKeyExists(t, config, "user_profiles")
			assertArrayItems(t, config["user_profiles"].([]interface{}), append(make([]interface{}, 0), "user_profile"))

			assertKeyExists(t, config, "report_configuration_ids")
			assertArrayItems(t, config["report_configuration_ids"].([]interface{}), append(make([]interface{}, 0), "report_configuration_id"))

			assertKeyExists(t, config, "accounts")
			assertArrayItems(t, config["accounts"].([]interface{}), append(make([]interface{}, 0), "account"))

			assertKeyExists(t, config, "fields")
			assertArrayItems(t, config["fields"].([]interface{}), append(make([]interface{}, 0), "field"))

			assertKeyExists(t, config, "breakdowns")
			assertArrayItems(t, config["breakdowns"].([]interface{}), append(make([]interface{}, 0), "breakdown"))

			assertKeyExists(t, config, "action_breakdowns")
			assertArrayItems(t, config["action_breakdowns"].([]interface{}), append(make([]interface{}, 0), "action_breakdown"))

			assertKeyExists(t, config, "pages")
			assertArrayItems(t, config["pages"].([]interface{}), append(make([]interface{}, 0), "page"))

			assertKeyExists(t, config, "repositories")
			assertArrayItems(t, config["repositories"].([]interface{}), append(make([]interface{}, 0), "repository"))

			assertKeyExists(t, config, "dimension_attributes")
			assertArrayItems(t, config["dimension_attributes"].([]interface{}), append(make([]interface{}, 0), "dimension_attribute"))

			assertKeyExists(t, config, "columns")
			assertArrayItems(t, config["columns"].([]interface{}), append(make([]interface{}, 0), "column"))

			assertKeyExists(t, config, "manager_accounts")
			assertArrayItems(t, config["manager_accounts"].([]interface{}), append(make([]interface{}, 0), "manager_account"))

			assertKeyExists(t, config, "profiles")
			assertArrayItems(t, config["profiles"].([]interface{}), append(make([]interface{}, 0), "profile"))

			assertKeyExists(t, config, "site_urls")
			assertArrayItems(t, config["site_urls"].([]interface{}), append(make([]interface{}, 0), "site_url"))

			assertKeyExists(t, config, "api_keys")
			assertArrayItems(t, config["api_keys"].([]interface{}), append(make([]interface{}, 0), "api_key"))

			assertKeyExists(t, config, "advertisers_id")
			assertArrayItems(t, config["advertisers_id"].([]interface{}), append(make([]interface{}, 0), "advertiser_id"))

			assertKeyExists(t, config, "hosts")
			assertArrayItems(t, config["hosts"].([]interface{}), append(make([]interface{}, 0), "host"))

			assertKeyExists(t, config, "advertisers")
			assertArrayItems(t, config["advertisers"].([]interface{}), append(make([]interface{}, 0), "advertiser"))

			assertKeyExists(t, config, "organizations")
			assertArrayItems(t, config["organizations"].([]interface{}), append(make([]interface{}, 0), "organization"))

			assertKeyExists(t, config, "account_ids")
			assertArrayItems(t, config["account_ids"].([]interface{}), append(make([]interface{}, 0), "account_id"))

			assertKeyExists(t, config, "packed_mode_tables")
			assertArrayItems(t, config["packed_mode_tables"].([]interface{}), append(make([]interface{}, 0), "packed_mode_table"))

			assertKeyExists(t, config, "adobe_analytics_configurations")

			adobe_analytics_configurations := config["adobe_analytics_configurations"].([]interface{})

			assertEqual(t, len(adobe_analytics_configurations), 1)

			adobe_analytics_configuration := adobe_analytics_configurations[0].(map[string]interface{})

			assertKeyExistsAndHasValue(t, adobe_analytics_configuration, "sync_mode", "sync_mode")

			assertKeyExists(t, adobe_analytics_configuration, "report_suites")
			assertArrayItems(t, adobe_analytics_configuration["report_suites"].([]interface{}), append(make([]interface{}, 0), "report_suite"))

			assertKeyExists(t, adobe_analytics_configuration, "elements")
			assertArrayItems(t, adobe_analytics_configuration["elements"].([]interface{}), append(make([]interface{}, 0), "element"))

			assertKeyExists(t, adobe_analytics_configuration, "metrics")
			assertArrayItems(t, adobe_analytics_configuration["metrics"].([]interface{}), append(make([]interface{}, 0), "metric"))

			assertKeyExists(t, adobe_analytics_configuration, "metrics")
			assertArrayItems(t, adobe_analytics_configuration["metrics"].([]interface{}), append(make([]interface{}, 0), "metric"))

			assertKeyExists(t, adobe_analytics_configuration, "calculated_metrics")
			assertArrayItems(t, adobe_analytics_configuration["calculated_metrics"].([]interface{}), append(make([]interface{}, 0), "calculated_metric"))

			assertKeyExists(t, adobe_analytics_configuration, "segments")
			assertArrayItems(t, adobe_analytics_configuration["segments"].([]interface{}), append(make([]interface{}, 0), "segment"))

			assertKeyExists(t, config, "reports")
			reports := config["reports"].([]interface{})
			assertEqual(t, len(reports), 1)
			report := reports[0].(map[string]interface{})

			assertKeyExistsAndHasValue(t, report, "table", "table")
			assertKeyExistsAndHasValue(t, report, "config_type", "config_type")
			assertKeyExistsAndHasValue(t, report, "prebuilt_report", "prebuilt_report")
			assertKeyExistsAndHasValue(t, report, "report_type", "report_type")
			assertKeyExistsAndHasValue(t, report, "filter", "filter")

			assertKeyExists(t, report, "fields")
			assertArrayItems(t, report["fields"].([]interface{}), append(make([]interface{}, 0), "field"))

			assertKeyExists(t, report, "dimensions")
			assertArrayItems(t, report["dimensions"].([]interface{}), append(make([]interface{}, 0), "dimension"))

			assertKeyExists(t, report, "metrics")
			assertArrayItems(t, report["metrics"].([]interface{}), append(make([]interface{}, 0), "metric"))

			assertKeyExists(t, report, "segments")
			assertArrayItems(t, report["segments"].([]interface{}), append(make([]interface{}, 0), "segment"))

			assertKeyExists(t, config, "custom_tables")
			custom_tables := config["custom_tables"].([]interface{})
			assertEqual(t, len(custom_tables), 1)
			custom_table := custom_tables[0].(map[string]interface{})

			assertKeyExistsAndHasValue(t, custom_table, "table_name", "table_name")
			assertKeyExistsAndHasValue(t, custom_table, "config_type", "config_type")
			assertKeyExistsAndHasValue(t, custom_table, "aggregation", "aggregation")
			assertKeyExistsAndHasValue(t, custom_table, "action_report_time", "action_report_time")
			assertKeyExistsAndHasValue(t, custom_table, "click_attribution_window", "click_attribution_window")
			assertKeyExistsAndHasValue(t, custom_table, "view_attribution_window", "view_attribution_window")
			assertKeyExistsAndHasValue(t, custom_table, "prebuilt_report_name", "prebuilt_report_name")

			assertKeyExists(t, custom_table, "fields")
			assertArrayItems(t, custom_table["fields"].([]interface{}), append(make([]interface{}, 0), "field"))
			assertKeyExists(t, custom_table, "breakdowns")
			assertArrayItems(t, custom_table["breakdowns"].([]interface{}), append(make([]interface{}, 0), "breakdown"))
			assertKeyExists(t, custom_table, "action_breakdowns")
			assertArrayItems(t, custom_table["action_breakdowns"].([]interface{}), append(make([]interface{}, 0), "action_breakdown"))

			assertKeyExists(t, config, "project_credentials")
			project_credentials := config["project_credentials"].([]interface{})
			assertEqual(t, len(project_credentials), 1)
			project_credential := project_credentials[0].(map[string]interface{})

			assertKeyExistsAndHasValue(t, project_credential, "project", "project")
			assertKeyExistsAndHasValue(t, project_credential, "api_key", "api_key")
			assertKeyExistsAndHasValue(t, project_credential, "secret_key", "secret_key")

			assertKeyExists(t, config, "secrets_list")
			secrets_list := config["secrets_list"].([]interface{})
			assertEqual(t, len(secrets_list), 1)
			function_secret := secrets_list[0].(map[string]interface{})

			assertKeyExistsAndHasValue(t, function_secret, "key", "key")
			assertKeyExistsAndHasValue(t, function_secret, "value", "value")

			connectorMockData = createMapFromJsonString(t, connectorMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceUpdate(t *testing.T) {
	mockClient.Reset()
	updateIteration := 0

	checkPatternNotRepresentedIfNotSet := func(t *testing.T, body map[string]interface{}) {
		assertKeyExists(t, body, "config")
		config := body["config"].(map[string]interface{})
		_, ok := config["pattern"]
		assertEqual(t, ok, false)
	}

	connectorMockUpdateGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			checkPatternNotRepresentedIfNotSet(t, requestBodyToJson(t, req))
			connectorMockData = createMapFromJsonString(t, connectorUpdateResponse1)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePatchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			updateIteration++
			checkPatternNotRepresentedIfNotSet(t, requestBodyToJson(t, req))
			connectorMockData = createMapFromJsonString(t, connectorUpdateResponse2)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceEmptyConfig(t *testing.T) {
	mockClient.Reset()

	connectorEmptyMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = createMapFromJsonString(t, connectorWithoutConfig)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func TestResourceConnectorConfigMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: connectorConfigMappingTfConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockPostHandler.Interactions, 1)
				assertEqual(t, connectorMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_sheets"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service_version", "1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "schedule_type", "auto"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.is_historical_sync", "true"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.update_state", "on_schedule"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.setup_state", "incomplete"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.sync_state", "paused"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.tasks.0.code", "task_code"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.tasks.0.message", "task_message"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.warnings.0.code", "warning_code"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.warnings.0.message", "warning_message"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "google_sheets_schema.table"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "sync_frequency", "5"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "paused", "true"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "pause_after_trial", "true"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),

			// check sensitive fields are have original values
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.oauth_token", "oauth_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.oauth_token_secret", "oauth_token_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.consumer_key", "consumer_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.client_secret", "client_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.private_key", "private_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.s3role_arn", "s3role_arn"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.ftp_password", "ftp_password"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.sftp_password", "sftp_password"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.api_key", "api_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.password", "password"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.secret_key", "secret_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.pem_certificate", "pem_certificate"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.access_token", "access_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.api_secret", "api_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.api_access_token", "api_access_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.secret", "secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.consumer_secret", "consumer_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.secrets", "secrets"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.api_token", "api_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.encryption_key", "encryption_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.pat", "pat"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.function_trigger", "function_trigger"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.token_key", "token_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.token_secret", "token_secret"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorMockDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceConnectorUpdateMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			sync_frequency = 5
			paused = true
			pause_after_trial = true
			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			config {
				user = "user"
				password = "password"
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				assertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			sync_frequency = 1440
			daily_sync_time = "3:30"
			paused = false
			pause_after_trial = false
			trust_certificates = true
			trust_fingerprints = true
			run_setup_tests = true

			config {
				user = "user1"
				password = "password1"
				host = "host"
				port = "123"
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				assertEqual(t, connectorMockUpdateGetHandler.Interactions, 4)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdate(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestResourceConnectorEmptyConfigMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			sync_frequency = 5
			paused = true
			pause_after_trial = true
			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false
			
			#config {}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorEmptyMockPostHandler.Interactions, 1)
				assertEqual(t, connectorEmptyMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceEmptyConfig(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorEmptyMockDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceConnectorListsConfigMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: connectorConfigListsMappingTfConfig,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorListsMockPostHandler.Interactions, 1)
				assertEqual(t, connectorListsMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceListMappingConfig(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorListsMockDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
