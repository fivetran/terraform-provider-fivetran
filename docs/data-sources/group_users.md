---
page_title: "Data Source: fivetran_group_users"
---

# Data Source: fivetran_group_users

This data source returns a list of information about all users within a group in your Fivetran account.

## Example

```hcl
data "fivetran_group_users" "group_users" {
        id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` - The unique identifier for the group within the Fivetran system.

### Read-Only

- `users` - see [below for nested schema](#nestedatt--users)

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `created_at` 
- `email` 
- `family_name` 
- `given_name` 
- `id` 
- `invited` 
- `logged_in_at` 
- `phone` 
- `picture` 
- `verified` 


