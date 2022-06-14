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

- `group_id` - The group id within the account.

### Optional

- `user` - Manages user assignment to a group (see [below for nested schema](#nestedblock--user))

### Read-Only

- `id` -  The resource `id`.

<a id="nestedblock--user"></a>
### Nested Schema for `user`

Required:

- `id` - The user id.
- `role` - The group role name that you would like to assign this user to. Available roles could be found at the [Roles](https://fivetran.com/account/roles) page on the Fivetran Dashboard.

## Import

To import an existing `fivetran_group_users` resource into your terraform state you need to get `Destination Group ID` on the `Destination` page at the Fivetran Dashboard.
To retrieve existing groups use [Data Source: fivetran_groups](/docs/data-sources/groups).
Then define an empty resource in your .tf configuration:

```hcl
resource "fivetran_group_users" "my_imported_fivetran_group_users" {

}
```

And call `terraform import` command:

```
terraform import fivetran_group_users.my_imported_fivetran_group_users <your Destination Group ID>
```

Then copy-paste values from the state to your .tf config, use `terraform state show`:

```
terraform state show 'fivetran_group_users.my_imported_fivetran_group_users'
```
