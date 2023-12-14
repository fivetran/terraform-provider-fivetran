---
page_title: "Data Source: fivetran_group_ssh_key"
---

# Data Source: fivetran_group_ssh_key

This data source returns public key from SSH key pair associated with the group.

## Example Usage

```hcl
data "fivetran_group_ssh_key" "my_group_public_key" {
    id = "group_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The unique identifier for the group within the Fivetran system.

### Read-Only

- `public_key` (String) Public key from SSH key pair associated with the group.