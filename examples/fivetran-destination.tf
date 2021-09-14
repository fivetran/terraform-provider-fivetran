
#################
### RESOURCES ###
#################

### fivetran_destination
resource "fivetran_destination" "my_dest1" {
    group_id = fivetran_group.my_new_group_1.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "US"
    trust_certificates = "true"
    trust_fingerprints = "true"
    run_setup_tests = "true"

    config {
        host = "terraform-pgsql-connector-test.cp0rdhwjbsae.us-east-1.rds.amazonaws.com"
        port = 5432
        user = "postgres"
        password = "zzzzzzzzzzzxzxzxzxzzx"
        database = "fivetran"
        connection_type = "Directly"
        # connection_type = "SSHTunnel"
        # tunnel_host = "my.tunnel.host.com"
        # tunnel_port = "3232"
        # tunnel_user = "usertunnel"
    }
}

####################
### DATA SOURCES ###
####################

# ### fivetran_destination
# data "fivetran_destination" "dest1" {
#     id = fivetran_destination.my_dest1.id
# }

# output "dest_output" {
#     value = data.fivetran_destination.dest1
# }
