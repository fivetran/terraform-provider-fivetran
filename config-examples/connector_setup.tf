terraform {
  required_providers {
    fivetran = {
        version = ">= 1.0.0"                            
        source = "fivetran/fivetran"
    }
  }
}

provider "fivetran" {

}

resource "fivetran_group" "group" {
    name = "MyGroup"
}

resource "fivetran_destination" "destination" {
    group_id = fivetran_group.group.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "GCP_US_EAST4"
    trust_certificates = "true"
    trust_fingerprints = "true"
    run_setup_tests = "true"
    networking_method = "Directly"

    config {
        host = "destination.host"
        port = 5432
        user = "postgres"
        password = "myPassword"
        database = "myDatabaseName"
    }
}

resource "fivetran_connector" "connector" {
    group_id = fivetran_group.group.id
    service = "fivetran_log"
    
    run_setup_tests = true

    destination_schema {
        name = "my_fivetran_log_connector"
    } 

    config {
        is_account_level_connector = "false"
    }

    depends_on = [
        fivetran_destination.destination
    ]
}

resource "fivetran_connector_schedule" "connector_schedule" {
    connector_id = fivetran_connector.connector.id

    paused = "false" 
    pause_after_trial = "false"
    sync_frequency = "60"
}