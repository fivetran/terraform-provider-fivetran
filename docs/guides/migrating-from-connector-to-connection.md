---
page_title: "Migrating from fivetran_connector to fivetran_connection + fivetran_connection_config"
---

# Migrating from fivetran_connector to fivetran_connection + fivetran_connection_config

This guide explains how to migrate your existing `fivetran_connector` resources to the new split architecture using `fivetran_connection` and `fivetran_connection_config`.

## Why Migrate?

The new split architecture provides several benefits over the legacy `fivetran_connector` resource:

### **Separation of Concerns**
- **Metadata** (service type, schema, networking) managed by `fivetran_connection`
- **Configuration** (host, port, database settings) managed by `fivetran_connection_config`
- **Credentials** (passwords, API keys) managed separately in `auth` field

### **Better Credential Management**
- Rotate credentials **without touching connection configuration**
- Update config **without exposing credentials**
- Easier integration with secrets management tools
- Reduced risk of accidental changes

### **Improved Workflow**
- Cleaner state file structure
- Better resource dependency management
- More granular updates and change tracking
- Easier to understand and troubleshoot

### **Enhanced Security**
- Credentials isolated in separate resource
- Support for secrets from AWS Secrets Manager, HashiCorp Vault, etc.
- Clear separation between public config and sensitive auth

## When to Migrate

- **Strongly Recommended**: New projects and infrastructure (use new resources from the start)
- **Good Timing**: When updating connector configuration or during maintenance windows
- **No Downtime**: Migration is safe and doesn't interrupt data syncing
- **Plan Carefully**: Production environments with complex dependencies

**Current Status**: The `fivetran_connector` resource is fully supported and will continue to work. The new split resources (`fivetran_connection` + `fivetran_connection_config`) offer additional benefits for credential management and configuration flexibility.

**What this means for you:**
- Both resource patterns are fully supported
- You can migrate at your convenience during normal maintenance windows
- New projects should consider using `fivetran_connection` + `fivetran_connection_config` from the start
- Migration provides immediate benefits (credential rotation, better security, cleaner state)
- Existing `fivetran_connector` resources continue to work without any changes required

## Before You Begin

### Prerequisites

1. **Backup your Terraform state**
   ```bash
   terraform state pull > pre-migration-state-$(date +%Y%m%d-%H%M%S).json
   ```

2. **Note your connector details**
   ```bash
   # Get connector ID
   terraform state show fivetran_connector.NAME | grep "^id "

   # Export for later use
   export CONNECTOR_ID="your_connector_id_here"
   ```

3. **Verify current state is clean**
   ```bash
   terraform plan
   # Should show: No changes. Your infrastructure matches the configuration.
   ```

4. **Review your connector configuration**
   - Identify all `config` block fields
   - Identify all `auth` block fields
   - Note any networking settings (proxy_agent_id, private_link_id, etc.)

### Expected Behavior

**Important**: After migration, you'll need to run `terraform apply` **once** before the state stabilizes. This is expected behavior.

**Why?**
- The Fivetran API doesn't return `config` and `auth` fields (they're write-only for security)
- Operational flags (`run_setup_tests`, etc.) aren't persisted
- First apply populates these from your Terraform configuration
- After that, state is stable with no drift

## Migration Steps

### Step 1: Create New Resource Definitions

Transform your existing `fivetran_connector` configuration into the new split format.

#### Before (fivetran_connector)

```hcl
resource "fivetran_group" "example" {
  name = "My Destination"
}

resource "fivetran_connector" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "my_schema"
  }

  config {
    host          = "db.example.com"
    port          = 5432
    database      = "mydb"
    user          = "fivetran_user"
    password      = var.db_password
    update_method = "XMIN"
  }

  run_setup_tests    = false
  trust_certificates = false
  trust_fingerprints = false
}
```

#### After (fivetran_connection + fivetran_connection_config)

```hcl
resource "fivetran_group" "example" {
  name = "My Destination"
}

resource "fivetran_connection" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "my_schema"
  }

  config = jsonencode({
    update_method = "XMIN"
  })

  run_setup_tests    = false
  trust_certificates = false
  trust_fingerprints = false
}

resource "fivetran_connection_config" "postgres" {
  connection_id = "CONNECTOR_ID_HERE"  # Replace with actual ID (see Step 3)

  config = jsonencode({
    host          = "db.example.com"
    port          = 5432
    database      = "mydb"
    user          = "fivetran_user"
    update_method = "XMIN"
  })

  auth = jsonencode({
    password = var.db_password
  })

  run_setup_tests    = false
  trust_certificates = false
  trust_fingerprints = false
}
```

**Key Changes**:
1. Configuration split into two resources
2. `config` and `auth` are now JSON-encoded strings
3. Connection details moved to `fivetran_connection_config.config`
4. Credentials moved to `fivetran_connection_config.auth`
5. Minimal config (update_method) can stay in `fivetran_connection.config`

### Step 2: Backup and Prepare

```bash
# 1. Backup your existing configuration
cp your-connector-config.tf your-connector-config.tf.bak

# 2. Get the connector ID
CONNECTOR_ID=$(terraform state show fivetran_connector.postgres | grep "^id " | awk '{print $3}' | tr -d '"')
echo "Connector ID: $CONNECTOR_ID"

# 3. Update your new configuration with the actual connector ID
# Replace "CONNECTOR_ID_HERE" in fivetran_connection_config with the actual ID
```

### Step 3: Remove Old Resource from State

**Critical**: This only removes the resource from Terraform state. The actual connector continues running in Fivetran.

```bash
# Remove old connector from state
terraform state rm fivetran_connector.postgres

# Verify removal
terraform state list | grep connector
# Should NOT show fivetran_connector.postgres
```

### Step 4: Import to New Resources

```bash
# Import to fivetran_connection
terraform import fivetran_connection.postgres $CONNECTOR_ID

# Import to fivetran_connection_config
terraform import fivetran_connection_config.postgres $CONNECTOR_ID

# Verify import
terraform state list
# Should show:
# - fivetran_connection.postgres
# - fivetran_connection_config.postgres
```

**Output**:
```
fivetran_connection.postgres: Import prepared!
fivetran_connection.postgres: Refreshing state... [id=connector_abc123]

Import successful!
```

### Step 5: Review and Apply Changes

```bash
# Check what Terraform plans to do
terraform plan
```

**Expected Output**:
```
Plan: 0 to add, 2 to change, 0 to destroy.

Changes:
# fivetran_connection.postgres will be updated in-place
  ~ config               = (known after apply)
  + run_setup_tests      = false
  + trust_certificates   = false
  + trust_fingerprints   = false

# fivetran_connection_config.postgres will be updated in-place
  + config             = jsonencode({...})
  + auth               = jsonencode({...})
  + run_setup_tests    = false
```

**This is expected!** These fields are write-only or operational flags that must be populated from your configuration.

```bash
# Apply the changes
terraform apply

# Expected: Apply complete! Resources: 0 added, 2 changed, 0 destroyed.
```

### Step 6: Verify Migration Success

```bash
# Verify no more changes needed
terraform plan
```

**Expected Output**:
```
No changes. Your infrastructure matches the configuration.
```

**Migration complete!** Your connector is now using the new split architecture.

### Step 7: Clean Up

```bash
# Remove old configuration file
rm your-connector-config.tf.bak

# Or rename it for reference
mv your-connector-config.tf.bak archive/legacy-connector.tf.bak
```

## Post-Migration: Using Your Migrated Resources

### Updating Connection Configuration

Now you can update connection details independently:

```hcl
resource "fivetran_connection_config" "postgres" {
  connection_id = "connector_abc123"

  config = jsonencode({
    host          = "new-db.example.com"  # Changed host
    port          = 5433                   # Changed port
    database      = "mydb"
    user          = "fivetran_user"
    update_method = "XMIN"
  })

  auth = jsonencode({
    password = var.db_password
  })

  run_setup_tests = false
}
```

```bash
terraform plan
# Shows: 1 to change (only connection_config updated)

terraform apply
```

### Rotating Credentials

Rotate credentials without touching connection configuration:

```hcl
resource "fivetran_connection_config" "postgres" {
  connection_id = "connector_abc123"

  config = jsonencode({
    host          = "db.example.com"
    port          = 5432
    database      = "mydb"
    user          = "new_fivetran_user"  # New username
    update_method = "XMIN"
  })

  auth = jsonencode({
    password = var.new_db_password  # New password
  })

  run_setup_tests = false
}
```

```bash
terraform plan
# Shows: 1 to change (auth updated)

terraform apply
```

### Using with Secrets Management

```hcl
# AWS Secrets Manager
data "aws_secretsmanager_secret_version" "db_creds" {
  secret_id = "prod/fivetran/postgres"
}

resource "fivetran_connection_config" "postgres" {
  connection_id = fivetran_connection.postgres.id

  config = jsonencode({
    host          = jsondecode(data.aws_secretsmanager_secret_version.db_creds.secret_string)["host"]
    port          = 5432
    database      = "mydb"
    user          = jsondecode(data.aws_secretsmanager_secret_version.db_creds.secret_string)["username"]
    update_method = "XMIN"
  })

  auth = jsonencode({
    password = jsondecode(data.aws_secretsmanager_secret_version.db_creds.secret_string)["password"]
  })

  run_setup_tests = false
}
```

## Migration Examples

### Example 1: PostgreSQL with Auth Block

#### Before
```hcl
resource "fivetran_connector" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "postgres_data"
  }

  config {
    host          = "db.example.com"
    port          = 5432
    database      = "production"
    update_method = "XMIN"
  }

  auth {
    user     = var.db_user
    password = var.db_password
  }

  run_setup_tests = false
}
```

#### After
```hcl
resource "fivetran_connection" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "postgres_data"
  }

  config = jsonencode({
    update_method = "XMIN"
  })

  run_setup_tests = false
}

resource "fivetran_connection_config" "postgres" {
  connection_id = fivetran_connection.postgres.id

  config = jsonencode({
    host          = "db.example.com"
    port          = 5432
    database      = "production"
    update_method = "XMIN"
  })

  auth = jsonencode({
    user     = var.db_user
    password = var.db_password
  })

  run_setup_tests = false
}
```

### Example 2: MySQL with Proxy Agent

#### Before
```hcl
resource "fivetran_connector" "mysql" {
  group_id = fivetran_group.example.id
  service  = "mysql"

  destination_schema {
    prefix = "mysql_data"
  }

  networking_method = "ProxyAgent"
  proxy_agent_id    = fivetran_proxy_agent.example.id

  config {
    host          = "mysql.internal.example.com"
    port          = 3306
    database      = "myapp"
    user          = "fivetran"
    password      = var.mysql_password
    update_method = "BINLOG"
  }

  run_setup_tests = false
}
```

#### After
```hcl
resource "fivetran_connection" "mysql" {
  group_id = fivetran_group.example.id
  service  = "mysql"

  destination_schema {
    prefix = "mysql_data"
  }

  networking_method = "ProxyAgent"
  proxy_agent_id    = fivetran_proxy_agent.example.id

  config = jsonencode({
    update_method = "BINLOG"
  })

  run_setup_tests = false
}

resource "fivetran_connection_config" "mysql" {
  connection_id = fivetran_connection.mysql.id

  config = jsonencode({
    host          = "mysql.internal.example.com"
    port          = 3306
    database      = "myapp"
    user          = "fivetran"
    update_method = "BINLOG"
  })

  auth = jsonencode({
    password = var.mysql_password
  })

  run_setup_tests = false
}
```

### Example 3: S3 with IAM Role (Auth Only)

#### Before
```hcl
resource "fivetran_connector" "s3" {
  group_id = fivetran_group.example.id
  service  = "s3"

  destination_schema {
    name = "s3_data"
  }

  config {
    bucket       = "my-data-bucket"
    prefix       = "fivetran/"
    role_arn     = aws_iam_role.fivetran.arn
  }

  run_setup_tests = false
}
```

#### After
```hcl
resource "fivetran_connection" "s3" {
  group_id = fivetran_group.example.id
  service  = "s3"

  destination_schema {
    name = "s3_data"
  }

  run_setup_tests = false
}

resource "fivetran_connection_config" "s3" {
  connection_id = fivetran_connection.s3.id

  config = jsonencode({
    bucket   = "my-data-bucket"
    prefix   = "fivetran/"
  })

  auth = jsonencode({
    role_arn = aws_iam_role.fivetran.arn
  })

  run_setup_tests = false
}
```


## Migrating Multiple Connectors

If you have multiple connectors in the same Terraform file, you can migrate them individually or together.

### Strategy 1: One at a Time (Recommended for Production)

Migrate each connector separately to minimize risk and make troubleshooting easier.

#### Before (Multiple Connectors)

```hcl
resource "fivetran_connector" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"
  
  destination_schema {
    prefix = "postgres_data"
  }
  
  config {
    host          = "db1.example.com"
    port          = 5432
    database      = "db1"
    user          = "user1"
    password      = var.postgres_password
    update_method = "XMIN"
  }
}

resource "fivetran_connector" "mysql" {
  group_id = fivetran_group.example.id
  service  = "mysql"
  
  destination_schema {
    prefix = "mysql_data"
  }
  
  config {
    host          = "db2.example.com"
    port          = 3306
    database      = "db2"
    user          = "user2"
    password      = var.mysql_password
    update_method = "BINLOG"
  }
}

resource "fivetran_connector" "s3" {
  group_id = fivetran_group.example.id
  service  = "s3"
  
  destination_schema {
    name = "s3_data"
  }
  
  config {
    bucket   = "my-bucket"
    role_arn = aws_iam_role.fivetran.arn
  }
}
```

#### Step-by-Step Migration

**Week 1: Migrate PostgreSQL connector**

```bash
# Get connector IDs
POSTGRES_ID=$(terraform state show fivetran_connector.postgres | grep "^id " | awk '{print $3}' | tr -d '"')
echo "PostgreSQL ID: $POSTGRES_ID"

# Remove from state
terraform state rm fivetran_connector.postgres

# Import to new resources
terraform import fivetran_connection.postgres $POSTGRES_ID
terraform import fivetran_connection_config.postgres $POSTGRES_ID

# Apply and verify
terraform apply
terraform plan  # Should show no changes for postgres, still shows old mysql and s3
```

**Week 2: Migrate MySQL connector**

```bash
# Get connector ID
MYSQL_ID=$(terraform state show fivetran_connector.mysql | grep "^id " | awk '{print $3}' | tr -d '"')

# Remove from state
terraform state rm fivetran_connector.mysql

# Import to new resources
terraform import fivetran_connection.mysql $MYSQL_ID
terraform import fivetran_connection_config.mysql $MYSQL_ID

# Apply and verify
terraform apply
terraform plan  # postgres and mysql stable, only s3 remains
```

**Week 3: Migrate S3 connector**

```bash
# Get connector ID
S3_ID=$(terraform state show fivetran_connector.s3 | grep "^id " | awk '{print $3}' | tr -d '"')

# Remove from state
terraform state rm fivetran_connector.s3

# Import to new resources
terraform import fivetran_connection.s3 $S3_ID
terraform import fivetran_connection_config.s3 $S3_ID

# Apply and verify
terraform apply
terraform plan  # All connectors migrated, no changes
```

#### After (All Migrated - One at a Time)

```hcl
# PostgreSQL (migrated week 1)
resource "fivetran_connection" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"
  
  destination_schema {
    prefix = "postgres_data"
  }
  
  config = jsonencode({
    update_method = "XMIN"
  })
}

resource "fivetran_connection_config" "postgres" {
  connection_id = "postgres_connector_id"  # From week 1
  
  config = jsonencode({
    host          = "db1.example.com"
    port          = 5432
    database      = "db1"
    user          = "user1"
    update_method = "XMIN"
  })
  
  auth = jsonencode({
    password = var.postgres_password
  })
}

# MySQL (migrated week 2)
resource "fivetran_connection" "mysql" {
  group_id = fivetran_group.example.id
  service  = "mysql"
  
  destination_schema {
    prefix = "mysql_data"
  }
  
  config = jsonencode({
    update_method = "BINLOG"
  })
}

resource "fivetran_connection_config" "mysql" {
  connection_id = "mysql_connector_id"  # From week 2
  
  config = jsonencode({
    host          = "db2.example.com"
    port          = 3306
    database      = "db2"
    user          = "user2"
    update_method = "BINLOG"
  })
  
  auth = jsonencode({
    password = var.mysql_password
  })
}

# S3 (migrated week 3)
resource "fivetran_connection" "s3" {
  group_id = fivetran_group.example.id
  service  = "s3"
  
  destination_schema {
    name = "s3_data"
  }
}

resource "fivetran_connection_config" "s3" {
  connection_id = "s3_connector_id"  # From week 3
  
  config = jsonencode({
    bucket = "my-bucket"
  })
  
  auth = jsonencode({
    role_arn = aws_iam_role.fivetran.arn
  })
}
```

### Strategy 2: Migrate All at Once

For development/testing environments, you can migrate multiple connectors in one session.

**Migration Script**:

```bash
#!/bin/bash
# migrate-all-connectors.sh

# Get all connector IDs
POSTGRES_ID=$(terraform state show fivetran_connector.postgres | grep "^id " | awk '{print $3}' | tr -d '"')
MYSQL_ID=$(terraform state show fivetran_connector.mysql | grep "^id " | awk '{print $3}' | tr -d '"')
S3_ID=$(terraform state show fivetran_connector.s3 | grep "^id " | awk '{print $3}' | tr -d '"')

echo "PostgreSQL: $POSTGRES_ID"
echo "MySQL: $MYSQL_ID"
echo "S3: $S3_ID"

# Update your new configuration file with these IDs before proceeding

# Remove all old connectors from state
terraform state rm fivetran_connector.postgres
terraform state rm fivetran_connector.mysql
terraform state rm fivetran_connector.s3

# Import all new resources
terraform import fivetran_connection.postgres $POSTGRES_ID
terraform import fivetran_connection_config.postgres $POSTGRES_ID

terraform import fivetran_connection.mysql $MYSQL_ID
terraform import fivetran_connection_config.mysql $MYSQL_ID

terraform import fivetran_connection.s3 $S3_ID
terraform import fivetran_connection_config.s3 $S3_ID

# Apply once for all connectors
terraform apply

# Verify
terraform plan  # Should show: No changes
```

### Mixed State: Old and New Resources Together

You can have both old and new resources in the same file during gradual migration:

```hcl
# Already migrated
resource "fivetran_connection" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"
  # ...
}

resource "fivetran_connection_config" "postgres" {
  connection_id = "postgres_id"
  # ...
}

# Not yet migrated (still using legacy)
resource "fivetran_connector" "mysql" {
  group_id = fivetran_group.example.id
  service  = "mysql"
  # ...
}

resource "fivetran_connector" "s3" {
  group_id = fivetran_group.example.id
  service  = "s3"
  # ...
}
```

This is completely valid and supported. The old and new resources don't interfere with each other.

### Best Practices for Multiple Connectors

**Do:**
- Backup state before starting migration
- Document which connectors have been migrated
- Test each migration in non-production first
- Keep a list of connector IDs for reference
- Migrate during maintenance windows
- Update one connector type at a time if similar configs

**Don't:**
- Migrate all production connectors at once (unless necessary)
- Skip state backups
- Forget to update connection_id with static values
- Mix up connector IDs between different connectors
- Rush the migration without testing

### Tracking Migration Progress

Create a checklist for multiple connectors:

```markdown
## Connector Migration Checklist

### PostgreSQL Connectors
- [ ] connector.postgres_prod (ID: abc123) - Scheduled: Week 1
- [ ] connector.postgres_staging (ID: def456) - Scheduled: Week 1
- [ ] connector.postgres_dev (ID: ghi789) - Scheduled: Week 1

### MySQL Connectors
- [ ] connector.mysql_prod (ID: jkl012) - Scheduled: Week 2
- [ ] connector.mysql_staging (ID: mno345) - Scheduled: Week 2

### S3 Connectors
- [ ] connector.s3_logs (ID: pqr678) - Scheduled: Week 3
- [ ] connector.s3_analytics (ID: stu901) - Scheduled: Week 3

### Status
- Completed: 0 / 7
- In Progress: 0
- Pending: 7
```

### Rollback Strategy for Multiple Connectors

If you need to rollback multiple connectors:

**Option 1: Rollback all at once**
```bash
# Restore pre-migration state
terraform state push pre-migration-state-backup.json

# Restore old config
mv connectors.tf.bak connectors.tf

# Verify
terraform plan  # Should show no changes
```

**Option 2: Rollback individual connectors**
```bash
# Rollback just one connector while keeping others migrated
# 1. Remove new resources from config file
# 2. Add old connector back to config file
# 3. Remove new resources from state
terraform state rm fivetran_connection.postgres
terraform state rm fivetran_connection_config.postgres

# 4. Re-import as old resource
terraform import fivetran_connector.postgres $POSTGRES_ID

# 5. Verify
terraform plan
```

## Troubleshooting

### Issue: Plan shows changes after import

**Symptom**:
```
Plan: 0 to add, 2 to change, 0 to destroy.
```

**Cause**: Write-only fields (`config`, `auth`) and operational flags not returned by API.

**Solution**: This is expected! Run `terraform apply` once to populate these fields:
```bash
terraform apply
terraform plan  # Now should show: No changes
```

### Issue: "Forces replacement" on connection_config

**Symptom**:
```
# fivetran_connection_config.postgres must be replaced
-/+ resource "fivetran_connection_config" "postgres" {
    ~ connection_id = "old_id" -> (known after apply) # forces replacement
```

**Cause**: Using a reference for `connection_id` instead of static value.

**Wrong**:
```hcl
connection_id = fivetran_connection.postgres.id  # Reference
```

**Correct** (during migration):
```hcl
connection_id = "connector_abc123"  # Static value
```

**Why**: The `connection_id` field has a "RequiresReplace" modifier. During migration, using a reference causes Terraform to see it as changing from the imported static value to a computed reference, triggering replacement.

**After migration stabilizes**, you can switch to using the reference if desired (though static is safer).

### Issue: Can't import config or auth fields

**Symptom**:
```
terraform state show fivetran_connection_config.postgres
# Shows only: connection_id and id
# Missing: config and auth
```

**Cause**: These are write-only fields for security. The API doesn't return them.

**Solution**: This is expected behavior. Specify `config` and `auth` in your Terraform configuration. They'll be populated on the first `terraform apply`.

### Issue: State file shows duplicate IDs

**Symptom**:
```bash
grep '"id"' terraform.tfstate | sort | uniq -c
   2 "id": "connector_abc123"
```

**Cause**: Both `fivetran_connection` and `fivetran_connection_config` share the same ID (the connector ID).

**Solution**: This is by design! Both resources manage different aspects of the same Fivetran connection. This is not a problem.

### Issue: Old connector still in state

**Symptom**:
```bash
terraform state list
fivetran_connector.postgres       # Old resource
fivetran_connection.postgres      # New resource
fivetran_connection_config.postgres
```

**Cause**: Forgot to run `terraform state rm` before import.

**Solution**:
```bash
# Remove old resource from state
terraform state rm fivetran_connector.postgres

# The connector continues running in Fivetran
```

### Issue: Credentials not updating

**Symptom**: Changed `auth.password` but connection still uses old credentials.

**Cause**: Typo in config or not applying changes.

**Solution**:
```bash
# 1. Verify your change
terraform plan
# Should show auth change

# 2. Apply the change
terraform apply

# 3. Verify in Fivetran UI that connection tests with new credentials
```

### Issue: Connection paused after migration

**Symptom**: Connection is paused and not syncing.

**Cause**: Not related to migration. Check connection setup state.

**Solution**:
```bash
# Check connection status in Fivetran UI
# Resume connection if needed
# Run setup tests: run_setup_tests = true
```

## How to Rollback

If you need to rollback the migration:

### Step 1: Restore Old State

```bash
# Restore from backup
terraform state push pre-migration-state-YYYYMMDD-HHMMSS.json

# Verify restoration
terraform state list
# Should show: fivetran_connector.postgres (old resource)
```

### Step 2: Restore Old Configuration

```bash
# Restore old Terraform configuration
mv your-connector-config.tf.bak your-connector-config.tf

# Remove new configuration
rm new-connection-config.tf  # or whatever you named it
```

### Step 3: Verify

```bash
terraform plan
# Should show: No changes (if state and config match)
```

**Note**: The connector never stopped running during migration or rollback. Only the Terraform state representation changed.

## Common Questions

### Q: Is there downtime during migration?

**A**: No! The connector continues running in Fivetran throughout the migration. You're only changing how Terraform tracks the resource.

### Q: Do I have to migrate?

**A**: No. The `fivetran_connector` resource is still fully supported. Migration is recommended for new projects and when you want the benefits of the split architecture.

### Q: Can I migrate back to fivetran_connector?

**A**: Yes, using the rollback procedure above. However, this is rarely needed.

### Q: What happens to my data during migration?

**A**: Nothing! Your data continues syncing normally. Migration only affects Terraform state, not the actual Fivetran connector.

### Q: Why do I need to apply after import?

**A**: The Fivetran API doesn't return write-only fields (`config`, `auth`) or operational flags for security. Terraform must populate these from your configuration.

### Q: Can I use a reference for connection_id after migration?

**A**: After the state stabilizes, you can switch from static ID to reference if you want. However, using the static ID is safer and avoids the RequiresReplace issue.

### Q: How do I know if migration was successful?

**A**: After running `terraform apply` once post-import, `terraform plan` should show "No changes."

### Q: What if I have multiple connectors?

**A**: You have two options:

**Option 1: Migrate one at a time (Recommended)**
- Safest approach for production
- Test each migration before moving to next
- Easier to rollback if issues occur
- Can spread across multiple maintenance windows

**Option 2: Migrate multiple at once**
- Faster for development/testing environments
- Follow same steps for each connector
- All connectors can coexist in same file

See the "Multiple Connectors" section below for detailed examples.

## Best Practices

### Do

- **Backup state** before migration
- **Test in non-production** first
- **Migrate one connector** at a time
- **Use static connection_id** during migration
- **Document your migration** for team reference
- **Verify** `terraform plan` shows no changes after migration
- **Use secrets management** for credentials in production

### Don't

- **Skip the backup** step
- **Migrate production** without testing first
- **Use references** for connection_id during initial migration
- **Expect zero drift** immediately after import (one apply is needed)
- **Panic** if first plan shows changes (this is expected)
- **Commit credentials** to version control

## Getting Help

If you encounter issues:

1. **Check this troubleshooting section** first
2. **Review Terraform plan output** carefully
3. **Verify connector ID** is correct
4. **Check state file** with `terraform state list` and `terraform state show`
5. **Consult Fivetran documentation**: [Fivetran REST API](https://fivetran.com/docs/rest-api/connectors)
6. **Open a support ticket** with Fivetran if needed

## See Also

- [`fivetran_connector`](../resources/connector.md) - Legacy connector resource
- [`fivetran_connection`](../resources/connection.md) - New connection resource
- [`fivetran_connection_config`](../resources/connection_config.md) - New config resource
- [Fivetran REST API Documentation](https://fivetran.com/docs/rest-api/connectors)
