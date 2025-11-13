# Testing the PrivateLink Fix

This document describes how to test the fix for the PrivateLink state inconsistency issue with Databricks destinations.

## Bug Description

When using PrivateLink with Databricks destinations, Fivetran's API returns modified values:
- `server_host_name`: Changed from the original Databricks hostname to the PrivateLink endpoint
- `cloud_provider`: Changed from "AZURE" to "AWS" (incorrect)

This caused Terraform to report: "Provider produced inconsistent result after apply"

## Fix Implementation

The fix preserves the user's original configuration values for `server_host_name` and `cloud_provider` when:
1. The destination `service` is "databricks"
2. The `networking_method` is "PrivateLink"

## Testing Locally

### Prerequisites
- Go 1.19 or later
- Access to a Fivetran account
- Azure Databricks instance with PrivateLink configured
- Fivetran PrivateLink setup

### Build the Provider

```bash
cd /home/jhiza/git/terraform-provider-fivetran
make build
```

### Test Configuration

Create a test Terraform configuration:

```hcl
terraform {
  required_providers {
    fivetran = {
      source  = "registry.terraform.io/fivetran/fivetran"
      version = "~> 1.9"
    }
  }
}

provider "fivetran" {
  # Configure your Fivetran API credentials
}

resource "fivetran_group" "test_group" {
  name = "test_privatelink_group"
}

resource "fivetran_destination" "test_destination" {
  group_id             = fivetran_group.test_group.id
  service              = "databricks"
  time_zone_offset     = "0"
  region               = "AZURE_EASTUS"
  trust_certificates   = true
  trust_fingerprints   = true
  daylight_saving_time_enabled = true
  run_setup_tests      = false
  networking_method    = "PrivateLink"
  private_link_id      = "<your_private_link_id>"

  config {
    auth_type             = "PERSONAL_ACCESS_TOKEN"
    catalog               = "<your_catalog>"
    server_host_name      = "<your_azure_databricks_hostname>"
    port                  = 443
    http_path             = "<your_http_path>"
    cloud_provider        = "AZURE"
    personal_access_token = "<your_token>"
  }
}
```

### Test Steps

1. Run `terraform plan` - should show resource creation
2. Run `terraform apply` - should succeed WITHOUT the "Provider produced inconsistent result" error
3. Run `terraform plan` again - should show "No changes" (not showing drift for `server_host_name` or `cloud_provider`)
4. Modify another field (e.g., `time_zone_offset`)
5. Run `terraform apply` - should succeed and preserve the original values

### Expected Behavior

**Before the fix:**
- `terraform apply` fails with: "Provider produced inconsistent result after apply"
- User needs to add `lifecycle { ignore_changes = [config.server_host_name, config.cloud_provider] }`

**After the fix:**
- `terraform apply` succeeds
- State file contains the user's original `server_host_name` and `cloud_provider` values
- No need for `lifecycle.ignore_changes` workaround

## Alternative Testing

If you don't have a full PrivateLink setup, you can:

1. Review the code changes in `fivetran/framework/resources/destination.go`
2. Check that the `preservePrivateLinkPlanValues()` function:
   - Only activates when `networking_method == "PrivateLink"` and `service == "databricks"`
   - Correctly preserves plan values for `server_host_name` and `cloud_provider`
   - Doesn't affect other destination types or networking methods

## Files Changed

- `fivetran/framework/resources/destination.go`: Added fix logic
- `CHANGELOG.md`: Documented the fix

