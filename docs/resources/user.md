---
page_title: "Resource: fivetran_user"
---

# Resource: fivetran_user

This resource allows you to create, update, and delete users.

## Example

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

### Read-Only

- `created_at` 
- `id` 
- `invited` 
- `last_updated` 
- `logged_in_at` 
- `verified` 
