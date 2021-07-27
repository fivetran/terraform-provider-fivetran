---
page_title: "Data Source: fivetran_groups"
---

# Data Source: fivetran_groups

This data source returns a list of all groups within your Fivetran account.

## Example

```hcl
data "fivetran_groups" "all" {
}
```

## Schema

### Read-Only

- `groups` - see [below for nested schema](#nestedatt--groups)

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `created_at` 
- `id` 
- `name` 
