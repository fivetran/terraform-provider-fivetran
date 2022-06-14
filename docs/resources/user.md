---
page_title: "Resource: fivetran_user"
---

# Resource: fivetran_user

This resource allows you to create, update, and delete users.

## Example Usage

```hcl
resource "fivetran_user" "user" {
    email = "user@email.address.com"
    given_name = "John"
    family_name = "Doe"
    phone = "+353 00 0000 0000"
}
```

## Schema

### Required

- `email` - The email address that the user has associated with their user profile.
- `given_name` - The first name of the user.
- `family_name` - The last name of the user.

### Optional

- `phone` - The phone number of the user.
- `picture` - The url of the user's avatar.
- `role` - The account role that you would like to assign this new user to. Possible values: Account Administrator, Account Billing, Account Analyst, Account Reviewer, Destination Creator, or a custom role with account-level permissions. You can find all roles at your (Roles)[https://fivetran.com/account/roles] page at Fivetran Dashboard.

### Read-Only

- `created_at` 
- `id` 
- `invited` 
- `last_updated` 
- `logged_in_at` 
- `verified` 

## Import

1. To import an existing `fivetran_user` resource into your Terraform state, you need to get `user_id`. 
You can retrieve all users using the [fivetran_users data source](/docs/data-sources/users).

2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_user" "my_imported_fivetran_user" {

}
```

3. Run the `terraform import` command:

```
terraform import fivetran_user.my_imported_fivetran_user <user_id>
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_user.my_imported_fivetran_user'
```
5. copy the values and paste them to your `.tf` configuration.
