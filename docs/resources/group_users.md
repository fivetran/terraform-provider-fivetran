---
page_title: "Resource: fivetran_group_users"
---

# Resource: fivetran_group_users

This resource allows you to create, update, and delete user memberships in groups.

## Example Usage

```hcl
resource "fivetran_group_users" "group_users" {
    group_id = fivetran_group.group.id

    user {
        email = "mail@example.com"
        role = "Destination Analyst"
    }

    user {
        email = "another_mail@example.com"
        role = "Destination Analyst"
    }
}
```

## Schema

### Required

- `group_id` - The group ID within the account.

### Optional

- `user` - Manages the user assignment to a group. See [Nested Schema for `user`](#nestedblock--user) for parameters used with nested schemas.

### Read-Only

- `id` -  The resource `id`.

<a id="nestedblock--user"></a>
### Nested Schema for `user`

Required:

- `id` - The user ID.
- `role` - The group role name that you would like to assign this user to. You can see the available roles on the [**Roles** tab](https://fivetran.com/account/roles) of the account management page in your Fivetran dashboard.

## Import

1. To import an existing `fivetran_group_users` resource into your Terraform state, you need to get **Destination Group ID** on the destination page in your Fivetran dashboard.
To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).
2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_group_users" "my_imported_fivetran_group_users" {

}
```

3. Run the `terraform import` command:

```
terraform import fivetran_group_users.my_imported_fivetran_group_users {your Destination Group ID}
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_group_users.my_imported_fivetran_group_users'
```
5. Copy the values and paste them to your `.tf` configuration.
