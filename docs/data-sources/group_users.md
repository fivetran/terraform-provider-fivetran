---
page_title: "Data Source: fivetran_group_users"
---

# Data Source: fivetran_group_users

This data source returns a list of information about all users within a group in your Fivetran account.

## Example Usage

```hcl
data "fivetran_group_users" "group_users" {
        id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` (String) The unique identifier for the user within the account.

### Read-Only

- `users` (Set of Object) (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `created_at` (String)
- `email` (String)
- `family_name` (String)
- `given_name` (String)
- `id` (String)
- `invited` (Boolean)
- `logged_in_at` (String)
- `phone` (String)
- `picture` (String)
- `role` (String)
- `verified` (Boolean)
