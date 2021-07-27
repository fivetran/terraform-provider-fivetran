---
page_title: "Data Source: fivetran_user"
---

# Data Source: fivetran_user

This data source returns a user object.

## Example

```hcl
data "fivetran_user" "my_user" {
    id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` - The unique identifier for the user within the Fivetran system.

### Read-Only

- `created_at` 
- `email` 
- `family_name` 
- `given_name` 
- `invited` 
- `logged_in_at` 
- `phone` 
- `picture` 
- `verified` 
