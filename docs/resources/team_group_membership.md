---
page_title: "Resource: fivetran_team_group_membership"
---

# Resource: fivetran_team_group_membership

This resource allows you to create, update, and delete group membership for teams

## Example Usage

```hcl
resource "fivetran_team_group_membership" "test_team_group_membership" {
    provider = fivetran-provider

    team_id = "test_team"

    group {
        connector_id = "test_connector"
        group_id = "test_group"
        role = "Destination Administrator"
    }

    group {
        connector_id = "test_connector"
        group_id = "test_group"
        role = "Destination Administrator"
    }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `team_id` (String) The unique identifier for the team within your account.

### Optional

- `group` (Block Set) (see [below for nested schema](#nestedblock--group))

### Read-Only

- `id` (String) The unique identifier for resource.

<a id="nestedblock--group"></a>
### Nested Schema for `group`

Required:

- `group_id` (String) The group unique identifier
- `role` (String) The team's role that links the team and the group

Read-Only:

- `created_at` (String) The date and time the membership was created

## Import

1. To import an existing `fivetran_team_group_membership` resource into your Terraform state, you need to get `team_id` and `group_id`
You can retrieve all teams using the [fivetran_teams data source](/docs/data-sources/teams).

2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_team_group_membership" "my_imported_fivetran_team_group_membership" {

}
```

3. Run the `terraform import` command:

```
terraform import fivetran_team_group_membership.my_imported_fivetran_team_group_membership {team_id}
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_team_group_membership.my_imported_fivetran_team_group_membership'
```
5. Copy the values and paste them to your `.tf` configuration.
