---
page_title: "Data Source: fivetran_users"
---

# Data Source: fivetran_users

This data source returns a list of all users within your Fivetran account.

## Example Usage

```hcl
data "fivetran_users" "users" {
}
```

## Schema

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


