---
page_title: "Data Source: fivetran_group"
---

# Data Source: fivetran_group

This data source returns a group object.

## Example

```hcl
data "fivetran_group" "my_group" {
    id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` - The unique identifier for the group within the Fivetran system.

### Read-Only

- `created_at`
- `name`
