#################
### RESOURCES ###
#################

# ### fivetran_group
resource "fivetran_group" "my_new_group_1" {
    name = "MyNewGroupAPITest1NEW1235791"

    user {
        id = "surveyor_reconvene"
        role = "ReadOnly"
    }

    user {
        id = fivetran_user.usertest1.id
        role = "ReadOnly"
    }

}

# output "grouptest1_output" {
#     value = fivetran_group.my_new_group_1
# }

####################
### DATA SOURCES ###
####################

# ### fivetran_group_connectors

# ### fivetran_groups
# data "fivetran_groups" "all" {}

# output "all_groups" {
#     value = data.fivetran_groups.all
# }

# ### fivetran_group
# data "fivetran_group" "my_group" {
#     id = fivetran_group.my_new_group_1.id
# }

# output "my_group_output" {
#     value = data.fivetran_group.my_group
# }

# ### fivetran_group_users
# data "fivetran_group_users" "my_group_users" {
#         # id = "rescuer_donator"
#         id = fivetran_group.my_new_group_1.id

# }

# output "my_group_users_output" {
#     value = data.fivetran_group_users.my_group_users
# }


