#################
### RESOURCES ###
#################
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

# ### fivetran_connectors_metadata
# data "fivetran_connectors_metadata" "sources" {
# }

# output "sources_output" {
#     value = data.fivetran_connectors_metadata.sources
# }
