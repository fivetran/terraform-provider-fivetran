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

    # user {
    #     id = fivetran_user.usertest12.id
    #     role = "ReadOnly"
    # }

    # user {
    #     id = "repaying_tangential"
    #     role = "ReadOnly"
    # }

    # user {
    #     id = "stiffness_grocery"
    #     role = "ReadOnly"
    # }
}

# output "grouptest1_output" {
#     value = fivetran_group.my_new_group_1
# }

####################
### DATA SOURCES ###
####################

# ### fivetran_group_connectors
# data "fivetran_group_connectors" "connector" {
#     id = "17z306ouk5cey"
#     schema = "webhooks.salame_audit"
# }

# output "group_connectors" {
#     value = data.fivetran_group_connectors.connector
# }

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

###############
### IMPORTS ###
###############

# resource "fivetran_user" "collateral_imputation" {
#     email        = "felipe.neuwald+apitest999@fivetran.com"
#     family_name  = "Neuwald API test 999 TERRAFORM"
#     given_name   = "Felipe"
# }

# output "user_collateral_imputation" {
#   value = fivetran_user.collateral_imputation
# }
