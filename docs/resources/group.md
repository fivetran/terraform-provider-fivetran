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

To import an existing `fivetran_group` resource into your terraform state you need to get `Destination Group ID` on the `Destination` page at the Fivetran Dashboard.
To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).
2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_group" "my_imported_fivetran_group" {

}
```

3. Call the `terraform import` command:

```
terraform import fivetran_group.my_imported_fivetran_group <your Destination Group ID>
```

4. Copy and paste the values from the state to your `.tf` configuration using the `terraform state show` command:

```
terraform state show 'fivetran_group.my_imported_fivetran_group'
```
