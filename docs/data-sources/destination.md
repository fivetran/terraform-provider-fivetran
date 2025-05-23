---
page_title: "Data Source: fivetran_destination"
---

# Data Source: fivetran_destination

This data source returns a destination object.

## Example Usage

```hcl
data "fivetran_destination" "dest" {
    id = "anonymous_mystery"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The unique identifier for the destination within the Fivetran system.

### Read-Only

- `config` (Block, Read-only) (see [below for nested schema](#nestedblock--config))
- `daylight_saving_time_enabled` (Boolean) Shift my UTC offset with daylight savings time (US Only)
- `group_id` (String) The unique identifier for the Group within the Fivetran system.
- `hybrid_deployment_agent_id` (String) The hybrid deployment agent ID that refers to the controller created for the group the connection belongs to. If the value is specified, the system will try to associate the connection with an existing agent.
- `networking_method` (String) Possible values: Directly, SshTunnel, ProxyAgent.
- `private_link_id` (String) The private link ID.
- `region` (String) Data processing location. This is where Fivetran will operate and run computation on data.
- `service` (String) The destination type id within the Fivetran system.
- `setup_status` (String) Destination setup status.
- `time_zone_offset` (String) Determines the time zone for the Fivetran sync schedule.

<a id="nestedblock--config"></a>
### Nested Schema for `config`

Read-Only:

- `always_encrypted` (Boolean) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `aurora_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_postgres_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_data_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_database`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_managed_db_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_rds_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_rds_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `panoply`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `periscope_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_gcp_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_rds_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `redshift`: Require TLS through Tunnel
	- Service `sql_server_rds_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_warehouse`: Specifies whether TLS is required. Must be populated if `connection_type` is set to `SshTunnel`.
- `auth` (String) Field usage depends on `service` value: 
	- Service `snowflake`: Password-based or key-based authentication type
- `auth_type` (String) Field usage depends on `service` value: 
	- Service `adls`: Authentication type
	- Service `databricks`: Authentication type
	- Service `new_s3_datalake`: Authentication type
	- Service `onelake`: Authentication type
	- Service `redshift`: Authentication type. Default value: `PASSWORD`.
- `aws_access_key_id` (String) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: AWS access key to access the S3 bucket and AWS Glue
- `aws_secret_access_key` (String, Sensitive) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: AWS secret access key to access the S3 bucket and AWS Glue
- `bootstrap_servers` (Set of String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Comma-separated list of Confluent Cloud servers in the `server:port` format.
- `bucket` (String) Field usage depends on `service` value: 
	- Service `big_query`: Customer bucket. If specified, your GCS bucket will be used to process the data instead of a Fivetran-managed bucket. The bucket must be present in the same location as the dataset location.
	- Service `big_query_dts`: Customer bucket. If specified, your GCS bucket will be used to process the data instead of a Fivetran-managed bucket. The bucket must be present in the same location as the dataset location.
	- Service `managed_big_query`: Customer bucket. If specified, your GCS bucket will be used to process the data instead of a Fivetran-managed bucket. The bucket must be present in the same location as the dataset location.
	- Service `new_s3_datalake`: (Immutable) The name of the bucket to be used as destination
- `catalog` (String) Field usage depends on `service` value: 
	- Service `adls`: Catalog name
	- Service `databricks`: Catalog name
	- Service `new_s3_datalake`: Catalog name
	- Service `onelake`: Catalog name
- `client_id` (String) Field usage depends on `service` value: 
	- Service `adls`: Client id of service principal
	- Service `onelake`: Client ID of service principal
- `cloud_provider` (String) Field usage depends on `service` value: 
	- Service `databricks`: Databricks deployment cloud
- `cluster_id` (String) Field usage depends on `service` value: 
	- Service `panoply`: Cluster ID.
	- Service `periscope_warehouse`: Cluster ID.
	- Service `redshift`: Cluster ID. Must be populated if `connection_type` is set to `SshTunnel` and `auth_type` is set to `IAM`.
- `cluster_region` (String) Field usage depends on `service` value: 
	- Service `panoply`: Cluster region.
	- Service `periscope_warehouse`: Cluster region.
	- Service `redshift`: Cluster region. Must be populated if `connection_type` is set to `SshTunnel` and `auth_type` is set to `IAM`.
- `connection_method` (String)
- `connection_type` (String) Field usage depends on `service` value: 
	- Service `adls`: Connection method. Default value: `Directly`.
	- Service `aurora_postgres_warehouse`: Connection method. Default value: `Directly`.
	- Service `aurora_warehouse`: Connection method. Default value: `Directly`.
	- Service `azure_postgres_warehouse`: Connection method. Default value: `Directly`.
	- Service `azure_sql_data_warehouse`: Connection method. Default value: `Directly`.
	- Service `azure_sql_database`: Connection method. Default value: `Directly`.
	- Service `azure_sql_managed_db_warehouse`: Connection method. Default value: `Directly`.
	- Service `databricks`: Connection method. Default value: `Directly`.
	- Service `maria_rds_warehouse`: Connection method. Default value: `Directly`.
	- Service `maria_warehouse`: Connection method. Default value: `Directly`.
	- Service `mysql_rds_warehouse`: Connection method. Default value: `Directly`.
	- Service `mysql_warehouse`: Connection method. Default value: `Directly`.
	- Service `panoply`: Connection method. Default value: `Directly`.
	- Service `periscope_warehouse`: Connection method. Default value: `Directly`.
	- Service `postgres_gcp_warehouse`: Connection method. Default value: `Directly`.
	- Service `postgres_rds_warehouse`: Connection method. Default value: `Directly`.
	- Service `postgres_warehouse`: Connection method. Default value: `Directly`.
	- Service `redshift`: Connection method. Default value: `Directly`.
	- Service `snowflake`: Connection method. Default value: `Directly`.
	- Service `sql_server_rds_warehouse`: Connection method. Default value: `Directly`.
	- Service `sql_server_warehouse`: Connection method. Default value: `Directly`.
- `container_name` (String) Field usage depends on `service` value: 
	- Service `adls`: (Immutable) Container to store delta table files
	- Service `onelake`: Workspace name to store delta table files
- `controller_id` (String)
- `create_external_tables` (Boolean) Field usage depends on `service` value: 
	- Service `databricks`: Whether to create external tables
- `data_format` (String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Confluent Cloud message format.
- `data_set_location` (String) Field usage depends on `service` value: 
	- Service `big_query`: Data location. Datasets will reside in this location.
	- Service `big_query_dts`: Data location. Datasets will reside in this location.
	- Service `managed_big_query`: Data location. Datasets will reside in this location.
- `database` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Database name
	- Service `aurora_warehouse`: Database name
	- Service `azure_postgres_warehouse`: Database name
	- Service `azure_sql_data_warehouse`: Database name
	- Service `azure_sql_database`: Database name
	- Service `azure_sql_managed_db_warehouse`: Database name
	- Service `maria_rds_warehouse`: Database name
	- Service `maria_warehouse`: Database name
	- Service `mysql_rds_warehouse`: Database name
	- Service `mysql_warehouse`: Database name
	- Service `panoply`: Database name
	- Service `periscope_warehouse`: Database name
	- Service `postgres_gcp_warehouse`: Database name
	- Service `postgres_rds_warehouse`: Database name
	- Service `postgres_warehouse`: Database name
	- Service `redshift`: Database name
	- Service `snowflake`: Database name
	- Service `sql_server_rds_warehouse`: Database name
	- Service `sql_server_warehouse`: Database name
- `databricks_connection_type` (String) Field usage depends on `service` value: 
	- Service `adls`: Databricks Connection method. Default value: `Directly`.
	- Service `new_s3_datalake`: Databricks Connection method. Default value: `Directly`.
	- Service `onelake`: Databricks Connection method. Default value: `Directly`.
- `enable_external_storage_for_unstructured_files` (Boolean)
- `enable_remote_execution` (Boolean)
- `enable_single_topic` (Boolean) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Populate all tables in a single topic.
- `external_id` (String) Field usage depends on `service` value: 
	- Service `aws_msk_wh`: Fivetran generated External ID
	- Service `panoply`: Fivetran generated External ID
	- Service `periscope_warehouse`: Fivetran generated External ID
	- Service `redshift`: Fivetran generated External ID
- `external_location` (String) Field usage depends on `service` value: 
	- Service `databricks`: External location to store Delta tables. Default value: `""`  (null). By default, the external tables will reside in the `/{schema}/{table}` path, and if you specify an external location in the `{externalLocation}/{schema}/{table}` path.
- `external_stage_storage_provider` (String)
- `external_storage_integration` (String)
- `external_storage_parent_folder_uri` (String)
- `fivetran_glue_role_arn` (String)
- `fivetran_msk_role_arn` (String)
- `fivetran_role_arn` (String) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: ARN of the role which you created with different required policy mentioned in our setup guide
- `host` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Server name
	- Service `aurora_warehouse`: Server name
	- Service `azure_postgres_warehouse`: Server name
	- Service `azure_sql_data_warehouse`: Server name
	- Service `azure_sql_database`: Server name
	- Service `azure_sql_managed_db_warehouse`: Server name
	- Service `maria_rds_warehouse`: Server name
	- Service `maria_warehouse`: Server name
	- Service `mysql_rds_warehouse`: Server name
	- Service `mysql_warehouse`: Server name
	- Service `panoply`: Server name
	- Service `periscope_warehouse`: Server name
	- Service `postgres_gcp_warehouse`: Server name
	- Service `postgres_rds_warehouse`: Server name
	- Service `postgres_warehouse`: Server name
	- Service `redshift`: Server name
	- Service `snowflake`: Server name
	- Service `sql_server_rds_warehouse`: Server name
	- Service `sql_server_warehouse`: Server name
- `http_path` (String) Field usage depends on `service` value: 
	- Service `adls`: HTTP path
	- Service `databricks`: HTTP path
	- Service `new_s3_datalake`: HTTP path
	- Service `onelake`: HTTP path
- `is_private_key_encrypted` (Boolean) Field usage depends on `service` value: 
	- Service `snowflake`: Indicates that a private key is encrypted. The default value: `false`. The field can be specified if authentication type is `KEY_PAIR`.
- `is_private_link_required` (Boolean) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: We use PrivateLink by default if your s3 bucket is in the same region as Fivetran. Turning on this toggle ensures that Fivetran always connects to s3 bucket over PrivateLink. Learn more in our [PrivateLink documentation](https://fivetran.com/docs/connectors/databases/connection-options#awsprivatelinkbeta).
- `is_redshift_serverless` (Boolean) Field usage depends on `service` value: 
	- Service `redshift`: Is your destination Redshift Serverless
- `lakehouse_guid` (String) Field usage depends on `service` value: 
	- Service `onelake`: (Immutable) OneLake lakehouse GUID
- `lakehouse_name` (String) Field usage depends on `service` value: 
	- Service `onelake`: (Immutable) Name of your lakehouse
- `msk_sts_region` (String)
- `num_of_partitions` (Number) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Number of partitions per topic.
- `oauth2_client_id` (String) Field usage depends on `service` value: 
	- Service `adls`: OAuth 2.0 client ID
	- Service `databricks`: OAuth 2.0 client ID
	- Service `new_s3_datalake`: OAuth 2.0 client ID
	- Service `onelake`: OAuth 2.0 client ID
- `oauth2_secret` (String, Sensitive) Field usage depends on `service` value: 
	- Service `adls`: OAuth 2.0 secret
	- Service `databricks`: OAuth 2.0 secret
	- Service `new_s3_datalake`: OAuth 2.0 secret
	- Service `onelake`: OAuth 2.0 secret
- `passphrase` (String, Sensitive) Field usage depends on `service` value: 
	- Service `snowflake`: In case private key is encrypted, you are required to enter passphrase that was used to encrypt the private key. The field can be specified if authentication type is `KEY_PAIR`.
- `password` (String, Sensitive) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Database user password
	- Service `aurora_warehouse`: Database user password
	- Service `azure_postgres_warehouse`: Database user password
	- Service `azure_sql_data_warehouse`: Database user password
	- Service `azure_sql_database`: Database user password
	- Service `azure_sql_managed_db_warehouse`: Database user password
	- Service `maria_rds_warehouse`: Database user password
	- Service `maria_warehouse`: Database user password
	- Service `mysql_rds_warehouse`: Database user password
	- Service `mysql_warehouse`: Database user password
	- Service `panoply`: Database user password
	- Service `periscope_warehouse`: Database user password
	- Service `postgres_gcp_warehouse`: Database user password
	- Service `postgres_rds_warehouse`: Database user password
	- Service `postgres_warehouse`: Database user password
	- Service `redshift`: Database user password. Required if authentication type is `PASSWORD`.
	- Service `snowflake`: Database user password. The field should be specified if authentication type is `PASSWORD`.
	- Service `sql_server_rds_warehouse`: Database user password
	- Service `sql_server_warehouse`: Database user password
- `personal_access_token` (String, Sensitive) Field usage depends on `service` value: 
	- Service `adls`: Personal access token
	- Service `databricks`: Personal access token
	- Service `new_s3_datalake`: Personal access token
	- Service `onelake`: Personal access token
- `port` (Number) Field usage depends on `service` value: 
	- Service `adls`: Server port number
	- Service `aurora_postgres_warehouse`: Server port number
	- Service `aurora_warehouse`: Server port number
	- Service `azure_postgres_warehouse`: Server port number
	- Service `azure_sql_data_warehouse`: Server port number
	- Service `azure_sql_database`: Server port number
	- Service `azure_sql_managed_db_warehouse`: Server port number
	- Service `databricks`: Server port number
	- Service `maria_rds_warehouse`: Server port number
	- Service `maria_warehouse`: Server port number
	- Service `mysql_rds_warehouse`: Server port number
	- Service `mysql_warehouse`: Server port number
	- Service `new_s3_datalake`: Server port number
	- Service `onelake`: Server port number
	- Service `panoply`: Server port number
	- Service `periscope_warehouse`: Server port number
	- Service `postgres_gcp_warehouse`: Server port number
	- Service `postgres_rds_warehouse`: Server port number
	- Service `postgres_warehouse`: Server port number
	- Service `redshift`: Server port number
	- Service `snowflake`: Server port number
	- Service `sql_server_rds_warehouse`: Server port number
	- Service `sql_server_warehouse`: Server port number
- `prefix_path` (String) Field usage depends on `service` value: 
	- Service `adls`: (Immutable) path/to/data within the container
	- Service `new_s3_datalake`: (Immutable) Prefix path of the bucket for which you have configured access policy. It is not required if access has been granted to entire Bucket in the access policy
	- Service `onelake`: (Immutable) path/to/data within your lakehouse inside the Files directory
- `private_key` (String, Sensitive) Field usage depends on `service` value: 
	- Service `snowflake`: Private access key.  The field should be specified if authentication type is `KEY_PAIR`.
- `project_id` (String) Field usage depends on `service` value: 
	- Service `big_query`: BigQuery project ID
- `public_key` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Public Key
	- Service `aurora_warehouse`: Public Key
	- Service `azure_postgres_warehouse`: Public Key
	- Service `azure_sql_data_warehouse`: Public Key
	- Service `azure_sql_database`: Public Key
	- Service `azure_sql_managed_db_warehouse`: Public Key
	- Service `maria_rds_warehouse`: Public Key
	- Service `maria_warehouse`: Public Key
	- Service `mysql_rds_warehouse`: Public Key
	- Service `mysql_warehouse`: Public Key
	- Service `panoply`: Public Key
	- Service `periscope_warehouse`: Public Key
	- Service `postgres_gcp_warehouse`: Public Key
	- Service `postgres_rds_warehouse`: Public Key
	- Service `postgres_warehouse`: Public Key
	- Service `redshift`: Public Key
	- Service `sql_server_rds_warehouse`: Public Key
	- Service `sql_server_warehouse`: Public Key
- `region` (String) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: Region of your AWS S3 bucket
- `registry_name` (String)
- `registry_sts_region` (String)
- `replication_factor` (Number) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Replication factor.
- `resource_id` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `aurora_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `azure_postgres_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `azure_sql_data_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `azure_sql_database`: Field to test Self serve Private Link
	- Service `azure_sql_managed_db_warehouse`: Field to test Self serve Private Link
	- Service `databricks`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `maria_rds_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `maria_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `mysql_rds_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `mysql_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `panoply`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `periscope_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `postgres_gcp_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `postgres_rds_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `postgres_warehouse`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `redshift`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `snowflake`: This field is currently being introduced to test the Self-serve Private Link functionality
	- Service `sql_server_rds_warehouse`: Field to test Self serve Private Link
	- Service `sql_server_warehouse`: Field to test Self serve Private Link
- `role` (String) Field usage depends on `service` value: 
	- Service `snowflake`: If not specified, Fivetran will use the user's default role
- `role_arn` (String, Sensitive) Field usage depends on `service` value: 
	- Service `redshift`: Role ARN with Redshift permissions. Required if authentication type is `IAM`.
- `sasl_mechanism` (String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Security protocol for Confluent Cloud interaction.
- `sasl_plain_key` (String, Sensitive) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Confluent Cloud SASL key.
- `sasl_plain_secret` (String, Sensitive) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Confluent Cloud SASL secret.
- `schema_compatibility` (String)
- `schema_registry` (String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Schema Registry
- `schema_registry_api_key` (String, Sensitive) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Schema registry API key.
- `schema_registry_api_secret` (String, Sensitive) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Schema registry API secret.
- `schema_registry_url` (String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Schema registry URL.
- `secret_key` (String, Sensitive) Field usage depends on `service` value: 
	- Service `big_query`: Private key of the customer service account. If specified, your service account will be used to process the data instead of the Fivetran-managed service account.
	- Service `big_query_dts`: Private key of the customer service account. If specified, your service account will be used to process the data instead of the Fivetran-managed service account.
	- Service `managed_big_query`: Private key of the customer service account. If specified, your service account will be used to process the data instead of the Fivetran-managed service account.
- `secret_value` (String, Sensitive) Field usage depends on `service` value: 
	- Service `adls`: Secret value for service principal
	- Service `onelake`: Secret value for service principal
- `security_protocol` (String) Field usage depends on `service` value: 
	- Service `confluent_cloud_wh`: Security protocol for Confluent Cloud interaction.
- `server_host_name` (String) Field usage depends on `service` value: 
	- Service `adls`: Server Host name
	- Service `databricks`: Server name
	- Service `new_s3_datalake`: Server host name
	- Service `onelake`: Server Host name
- `should_maintain_tables_in_databricks` (Boolean) Field usage depends on `service` value: 
	- Service `adls`: Should maintain tables in Databricks 
	- Service `new_s3_datalake`: Should maintain tables in Databricks 
	- Service `onelake`: Should maintain tables in Databricks
- `snapshot_retention_period` (String) Field usage depends on `service` value: 
	- Service `adls`: Snapshots older than the retention period are deleted every week. Default value: `ONE_WEEK`.
	- Service `new_s3_datalake`: Snapshots older than the retention period are deleted every week. Default value: `ONE_WEEK`.
	- Service `onelake`: Snapshots older than the retention period are deleted every week. Default value: `ONE_WEEK`.
- `snowflake_cloud` (String)
- `snowflake_region` (String)
- `storage_account_name` (String) Field usage depends on `service` value: 
	- Service `adls`: (Immutable) Storage account for Azure Data Lake Storage Gen2 name
	- Service `onelake`: (Immutable) Storage account for Azure Data Lake Storage Gen2 name
- `table_format` (String) Field usage depends on `service` value: 
	- Service `new_s3_datalake`: (Immutable) The table format in which you want to sync your tables. Valid values are ICEBERG and DELTA_LAKE
- `tenant_id` (String) Field usage depends on `service` value: 
	- Service `adls`: Tenant id of service principal
	- Service `onelake`: Tenant ID of service principal
- `tunnel_host` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `aurora_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_postgres_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_data_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_database`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_managed_db_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_rds_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_rds_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `panoply`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `periscope_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_gcp_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_rds_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `redshift`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_rds_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_warehouse`: SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.
- `tunnel_port` (Number) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `aurora_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_postgres_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_data_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_database`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_managed_db_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_rds_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_rds_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `panoply`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `periscope_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_gcp_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_rds_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `redshift`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_rds_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_warehouse`: SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.
- `tunnel_user` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `aurora_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_postgres_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_data_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_database`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `azure_sql_managed_db_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_rds_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `maria_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_rds_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `mysql_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `panoply`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `periscope_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_gcp_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_rds_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `postgres_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `redshift`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_rds_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
	- Service `sql_server_warehouse`: SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.
- `user` (String) Field usage depends on `service` value: 
	- Service `aurora_postgres_warehouse`: Database user name
	- Service `aurora_warehouse`: Database user name
	- Service `azure_postgres_warehouse`: Database user name
	- Service `azure_sql_data_warehouse`: Database user name
	- Service `azure_sql_database`: Database user name
	- Service `azure_sql_managed_db_warehouse`: Database user name
	- Service `maria_rds_warehouse`: Database user name
	- Service `maria_warehouse`: Database user name
	- Service `mysql_rds_warehouse`: Database user name
	- Service `mysql_warehouse`: Database user name
	- Service `panoply`: Database user name
	- Service `periscope_warehouse`: Database user name
	- Service `postgres_gcp_warehouse`: Database user name
	- Service `postgres_rds_warehouse`: Database user name
	- Service `postgres_warehouse`: Database user name
	- Service `redshift`: Database user name
	- Service `snowflake`: Database user name
	- Service `sql_server_rds_warehouse`: Database user name
	- Service `sql_server_warehouse`: Database user name
- `workspace_guid` (String) Field usage depends on `service` value: 
	- Service `onelake`: (Immutable) OneLake workspace GUID
- `workspace_name` (String) Field usage depends on `service` value: 
	- Service `onelake`: OneLake workspace name