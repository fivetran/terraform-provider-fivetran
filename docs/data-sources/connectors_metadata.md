---
page_title: "Data Source: fivetran_connectors_metadata"
---

# Data Source: fivetran_connectors_metadata

This data source returns all available source types within your Fivetran account. This data source makes it easier to display Fivetran connectors within your application because it provides metadata including the proper source name (‘Facebook Ad Account’ instead of facebook_ad_account), the source icon, and links to Fivetran resources. As we update source names and icons, that metadata will automatically update within this endpoint.

## Example Usage

```hcl
data "fivetran_connectors_metadata" "sources" {
}
```

## Schema

### Read-Only

- `sources` - see [below for nested schema](#nestedatt--sources)

<a id="nestedatt--sources"></a>
### Nested Schema for `sources`

Read-Only:

- `description` 
- `icon_url` 
- `id` 
- `link_to_docs` 
- `link_to_erd` 
- `name` 
- `type` 
