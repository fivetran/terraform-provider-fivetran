---
page_title: "Resource: fivetran_connector_certificates"
---

# Resource: fivetran_connector_certificates

This resource allows you to create, update, and delete connector certificates.

## Example Usage

```hcl
resource "fivetran_connector_certificates" "certificate" {
    provider = fivetran-provider

    certificate {
        hash = "jhgfJfgrI6yy..."
        encoded_cert = "encoded_cert"        
    }

    certificate {
        hash = "jhgfJfgrI6yy..."
        encoded_cert = "encoded_cert"        
    }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connector_id` (String) The unique identifier for the target connection within the Fivetran system.

### Optional

- `certificate` (Block Set) (see [below for nested schema](#nestedblock--certificate))

### Read-Only

- `id` (String) The unique identifier for the resource. Equal to target connection id.

<a id="nestedblock--certificate"></a>
### Nested Schema for `certificate`

Required:

- `encoded_cert` (String, Sensitive) Base64 encoded certificate.
- `hash` (String) Hash of the certificate.

Read-Only:

- `name` (String) Certificate name.
- `public_key` (String) The SSH public key.
- `sha1` (String) Certificate sha1.
- `sha256` (String) Certificate sha256.
- `type` (String) Type of the certificate.
- `validated_by` (String) User name who validated the certificate.
- `validated_date` (String) The date when certificate was approved.

## Import

1. To import an existing `fivetran_connector_certificates` resource into your Terraform state, you need to get **Fivetran Connector ID** on the **Setup** tab of the connector page in your Fivetran dashboard.

2. Retrieve all connectors in a particular group using the [fivetran_connectors data source](/docs/data-sources/connectors)

3. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_connector_certificates" "my_imported_connector_fingerprints" {

}
```

4. Run the `terraform import` command:

```
terraform import fivetran_connector_certificates.my_imported_connector_fingerprints {your Fivetran Connector ID}
```

5.  Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_connector_certificates.my_imported_connector_fingerprints'
```

6. Copy the values and paste them to your `.tf` configuration.