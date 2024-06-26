---
page_title: "Data Source: fivetran_external_logging"
---

# Data Source: fivetran_external_logging

This data source returns a logging service object.

## Example Usage

```hcl
data "fivetran_external_logging" "extlog" {
    id = "anonymous_mystery"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The unique identifier for the log service within the Fivetran system.

### Optional

- `config` (Block, Optional) (see [below for nested schema](#nestedblock--config))
- `run_setup_tests` (Boolean) Specifies whether the setup tests should be run automatically. The default value is TRUE.

### Read-Only

- `enabled` (Boolean) The boolean value specifying whether the log service is enabled.
- `group_id` (String) The unique identifier for the log service within the Fivetran system.
- `service` (String) The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.

<a id="nestedblock--config"></a>
### Nested Schema for `config`

Optional:

- `api_key` (String, Sensitive) API Key
- `channel` (String) Channel
- `enable_ssl` (Boolean) Enable SSL
- `external_id` (String) external_id
- `host` (String) Server name
- `hostname` (String) Server name
- `log_group_name` (String) Log Group Name
- `port` (Number) Port
- `primary_key` (String, Sensitive) Primary Key
- `project_id` (String) Project Id for Google Cloud Logging
- `region` (String) Region
- `role_arn` (String) Role Arn
- `sub_domain` (String) Sub Domain
- `token` (String, Sensitive) Token
- `workspace_id` (String) Workspace ID