#################
### RESOURCES ###
#################

### fivetran_connector
# resource "fivetran_connector" "my_conn1" {
#     group_id = fivetran_group.my_new_group_1.id
#     service = "google_sheets"
#     schema = "actually_this_is_the_connector_name"
#     sync_frequency = 5
#     paused = false
#     pause_after_trial = false

#     config {
#         table = "table"
#         sheet_id = "1Rmq_FN2kTNwWiT4adZKBxHBRmvfeBTIfKWi5B8ii9qk"
#         named_range = "range1"
#     }
# }

resource "fivetran_connector" "amplitude" {
    depends_on = [fivetran_destination.my_dest1]

    group_id = fivetran_group.my_new_group_1.id
    service = "amplitude"
    sync_frequency = 60
    paused = false
    pause_after_trial = false
    schema = "amplitude_connector"

    config {
        project_credentials {
            project = "1234"
            api_key = "asd"
            secret_key = "bbb"
        }

        project_credentials {
            project = "zzz"
            api_key = "zzzz111"
            secret_key = "zzzzcccc"
        }
    }
}

# resource "fivetran_connector" "pgsql" {
#     group_id = fivetran_group.my_new_group_1.id
#     service = "postgres_rds"
#     sync_frequency = 5
#     paused = false
#     pause_after_trial = false
#     schema = "my_pgsql_connector_schema"

#     config {
#         schema_prefix = "my_pgsql_connector_schema_prefix"
#         host = "terraform-pgsql-connector-test.cp0rdhwjbsae.us-east-1.rds.amazonaws.com"
#         port = "5432"
#         user = "postgres"
#         password = "zzzzzzzz"
#         database = "fivetran"
#     }
# }


# resource "fivetran_connector" "facebook_ads" {
#     group_id = fivetran_group.my_new_group_1.id
#     service = "facebook_ads"
#     schema = "facebook_ads_connector"
#     paused = false
#     pause_after_trial = false
#     sync_frequency = 60
#     # trust_certificates = true
#     # trust_fingerprints = true
#     # run_setup_tests = true

#     config {
#         sync_mode = "SpecificAccounts"
#         accounts = ["account123", "accountabc"]
#         timeframe_months = "THREE"
#         # custom_tables {
#         #     table_name = "table1"
#         #     config_type = "custom"
#         #     fields = ["field1", "field2", "field3"]
#         #     breakdowns = ["age", "gender"]
#         #     action_breakdowns = ["action_type", "action_device"]
#         #     aggregation = "week"
#         #     action_report_time = "impression"
#         #     click_attribution_window = "DAY_7"
#         #     view_attribution_window = "DAY 7"
#         # }

#         custom_tables {
#             table_name = "basic_ad_set"
#             config_type = "Prebuilt"
#             prebuilt_report_name = "BASIC_AD_SET"
#         }
#     }
# }

# resource "fivetran_connector" "adwords" {
#     group_id = fivetran_group.my_new_group_1.id
#     service = "adwords"
#     schema = "adwords_connector"
#     paused = false
#     pause_after_trial = false
#     sync_frequency = 60

#     config {
#         customer_id = "xxxx-xxxx-xxxxx"
#         sync_mode = "SpecificAccounts"
#         timeframe_months = "SIX"
#         accounts = []
#         manager_accounts = []
    
#         reports {
#             table = "table_1"
#             config_type = "Prebuilt"
#             prebuilt_report = "ACCOUNT_STATS"
#             fields = ["Impressions"]
#         }

#         reports {
#             table = "table_2"
#             config_type = "Prebuilt"
#             prebuilt_report = "ACCOUNT_STATS"
#             fields = ["Impressions", "Conversions"]            
#         }
#     }

#     auth {
#         client_access {
#             client_id = "my_client_id2"
#             client_secret = "my_client_secret"
#             user_agent = "my_company_name"
#             developer_token = "my_developer_token"
#         }
#         refresh_token = "my_refresh_token"
#     }
# }

####################
### DATA SOURCES ###
####################

# ### fivetran_connectors_metadata
# data "fivetran_connectors_metadata" "sources" {
# }

# output "sources_output" {
#     value = data.fivetran_connectors_metadata.sources
# }

# ### fivetran_connector
# data "fivetran_connector" "connector" {
#     id = fivetran_connector.amplitude.id
#     # id = "6dd8c23ddac06ed9445d7287d7758576"
#     # id = "audible_tion"
# }

# output "connector_output" {
#     value = data.fivetran_connector.connector
# }
