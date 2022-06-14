---
page_title: "Resource: fivetran_group"
---

# Resource: fivetran_group

This resource allows you to create, update, and delete groups.

## Example Usage

```hcl
resource "fivetran_group" "group" {
    name = "MyGroup"
}
```

## Schema

### Required

- `name` - The group name within the account. The name must start with a letter or underscore and can only contain letters, numbers, or underscores.

### Read-Only

- `created_at`
- `id`
- `last_updated`

## Import

1. To import an existing `fivetran_group` resource into your Terraform state, you need to get **Destination Group ID** on the destination page in your Fivetran dashboard.
To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).
2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_group" "my_imported_fivetran_group" {

}
```

3. Run the `terraform import` command:

```
terraform import fivetran_group.my_imported_fivetran_group <your Destination Group ID>
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_group.my_imported_fivetran_group'
```
