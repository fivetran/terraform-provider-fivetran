----
page_title: "Using schemas_json field"
subcategory: "Getting Started"
---

# How to set up Fivetran connector schema config using Terraform in `.json` format.

In cases when schema configuration is really big and you have to define more that 1000 tables settings it's better to set schema settings directly using `.json` file:

File `schema-config.json`:
```json
{
    "schema_0": {
        "enabled": true,
        "some_random_extra_field": "extra_value",
        "tables": {
            "table_0": {
                "some_random_extra_field": "extra_value",
                "enabled": true
            },
            ...
        }
    },
    "schema_2": {
        "enabled": true,
        "some_random_extra_field": "extra_value",
        "tables": {
            "table_0": {
                "some_random_extra_field": "extra_value",
                "enabled": true,
                "columns": {
                    "column_1": {
                        "enabled":  false
                    }
                }
            },
            ...
        }
    },
    ...
}
```

Configuration `.tf` file:
```hcl
resource "fivetran_connector_schema_config" "test_schema" {
    provider = fivetran-provider
    connector_id = "connector_id"
    schema_change_handling = "ALLOW_COLUMNS"
    schemas_json = file("path/to/schema-config.json")
}
```
Note: Enabled value should be of boolean type