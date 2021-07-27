---
page_title: "Resource: fivetran_group"
---

# Resource: fivetran_group

This resource allows you to create, update, and delete groups.

## Example Usage

```hcl
resource "fivetran_group" "group" {
    name = "MyGroup"

    user {
        id = "anonymous_mystery"
        role = "ReadOnly"
    }

    user {
        id = fivetran_user.user.id
        role = "ReadOnly"
    }
}
```

## Schema

### Required

- `name` - The group name within the account. The name must start with a letter or underscore and can only contain letters, numbers, or underscores.

### Optional

- `user` - Manages user assignment to a group (see [below for nested schema](#nestedblock--user))

### Read-Only

- `created_at`
- `creator`
- `id`
- `last_updated`

<a id="nestedblock--user"></a>
### Nested Schema for `user`

Required:

- `id` - The user id.
- `role` - The group role that you would like to assign this user to. Supported group roles: `ReadOnly`, `Uploader`, `Analyst`, `Admin`.
