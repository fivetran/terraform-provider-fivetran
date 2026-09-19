terraform {
    required_providers {
        fivetran = { 
            #version = "1.9.31"
            source = "fivetran/fivetran" 
        }
    }
}

provider "fivetran" {
    api_key    = "_moonbeam_acc_accountworthy_api_key_rbac"
    api_secret = "_moonbeam_acc_accountworthy_api_secret"
    api_url    = "http://localhost:8001/v1"
}

resource "fivetran_connector" "connector" {
    group_id = "_moonbeam"
    service = "itunes_connect"
    
    run_setup_tests = false

    destination_schema {
        name = "test_itunes_connect"
    } 

    config {
        issuer_id = "test_fivetran_itunes_connect_username"
    }
}

# resource "fivetran_connector_schedule" "test" {
#    connector_id = fivetran_connector.connector.id
#    sync_frequency  = "60"
#    paused          = "false"
#    pause_after_trial = "false"
# }

resource "fivetran_connector_schedule" "test" {
   connector_id = fivetran_connector.connector.id
   schedule {
     schedule_type = "INTERVAL"
     interval      = 60
   }
}

