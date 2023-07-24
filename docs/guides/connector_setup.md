----
page_title: "Initial Setup"
subcategory: "Getting Started"
---

# How to set up your Fivetran environment using Terraform

In this guide, we will set up a simple pipeline with one source using Fivetran Terraform Provider.

## Provider setup

First of all, you need to get your [Fivetran API Key and Secret](https://fivetran.com/docs/rest-api/getting-started#gettingstarted) and save it into environment variables:

```bash
export FIVETRAN_APIKEY=<your_Fivetran_API_key>
export FIVETRAN_APISECRET=<your_Fivetran_API_secret>
```

```hcl
# Terraform 0.13+ uses the Terraform Registry:

terraform {
  required_providers {
    fivetran = {
        version = "0.7.2"                                                   
        source = "fivetran/fivetran"
    }
  }
}

# Configure the Fivetran provider
provider "fivetran" {
#   We recommend to use environment variables `FIVETRAN_APIKEY` and `FIVETRAN_APISECRET` instead of explicit assignment
#   api_key = var.fivetran_api_key
#   api_secret = var.fivetran_api_secret
}

# Terraform 0.12- can be specified as:

# Configure the Fivetran provider
# provider "fivetran" {
#   api_key = "${var.fivetran_api_key}"
#   api_secret = "${var.fivetran_api_secret}"
# }
```

## Add your group and destination

The root resource for your Fivetran infrastructure setup is always `Destination group`. First of all, you need to set up the group:

```hcl
resource "fivetran_group" "group" {
    name = "MyGroup"
}
```

Once you have created the group, you need to associate a `Destination` with it:

```hcl
resource "fivetran_destination" "destination" {
    group_id = fivetran_group.group.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "GCP_US_EAST4"
    trust_certificates = "true"
    trust_fingerprints = "true"
    run_setup_tests = "true"

    config {
        host = "destination.host"
        port = 5432
        user = "postgres"
        password = "myPassword"
        database = "myDatabaseName"
        connection_type = "Directly"
    }
}
```

## Add your first connector

We are now ready to set up our first connector:

```hcl
resource "fivetran_connector" "connector" {
    group_id = fivetran_group.group.id
    service = "fivetran_log"
    run_setup_tests = true

    destination_schema {
        name = "my_fivetran_log_connector"
    } 

    config {
        is_account_level_connector = "false"
    }

    depends_on = [
        fivetran_destination.destination
    ]
}
```

## Configure connector schedule

We should configure how connector will be scheduled to sync:

```hcl
resource "fivetran_connector_schedule" "connector_schedule" {
    connector_id = fivetran_connector.connector.id
    sync_frequency = 60
    paused = false 
    pause_after_trial = false
}
```

Now we are ready to apply our configuration:

```bash
terraform apply
```

## Example configuration

An example .tf file with the configuration could be found [here](https://github.com/fivetran/terraform-provider-fivetran/tree/main/config-examples/connector_setup.tf).