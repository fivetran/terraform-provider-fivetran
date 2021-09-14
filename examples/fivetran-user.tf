#################
### RESOURCES ###
#################

### fivetran_user
resource "fivetran_user" "usertest1" {
    email = "felipe.neuwald+apitest.123579@fivetran.com"
    family_name = "Neuwald API test 999 terraform"
    given_name = "Felipe 123"
    phone = "+353 83 346 6015"
    # picture = "https://myPicturecom"
}

# resource "fivetran_user" "usertest12" {
#     email = "felipe.neuwald+apitest.12345@fivetran.com"
#     family_name = "Neuwald API test 999 terraform"
#     given_name = "Felipe 123"
#     phone = "+353 83 346 6015"
#     picture = "https://myPicturecom.com"
# }

# resource "fivetran_user" "usertest2" {
#     email = "felipe.neuwald+apitest5555@fivetran.com"
#     family_name = "Neuwald API test 999 terraform"
#     given_name = "Felipe User"
#     # phone = "+353 83 346 6015"
# }

# output "usertest2_output" {
#     value = fivetran_user.usertest2
# }

####################
### DATA SOURCES ###
####################

# ### fivetran_users
# data "fivetran_users" "users" {
# }

# output "users_output" {
#     value = data.fivetran_users.users
# }

# ### fivetran_user
# data "fivetran_user" "my_user" {
#     id = fivetran_user.usertest1.id
# }

# output "my_user_output" {
#     value = data.fivetran_user.my_user
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
