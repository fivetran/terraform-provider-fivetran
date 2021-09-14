# Fivetran Terraform Provider Examples

Configuration in this directory creates set of Fivetran resources.

The examples cover the creation of Fivetran connectors, destinations, users, and groups 

## Usage

Authenticate with Fivetran by setting the following environment variables

```bash
$ export FIVETRAN_APIKEY=YOUR_API_KEY
$ export FIVETRAN_APISECRET=YOUR_API_SECRET
```

To run these examples you need to execute:

```bash
$ terraform init
$ terraform plan
$ terraform apply
```

Run `terraform destroy` when you don't need these resources.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= x.x.x |

## Providers

No providers.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_fivetran"></a> [fivetran](#module\_vpc) | ../../ | 0.1.0  |

## Resources

No resources.

## Inputs

No inputs.

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_connector"></a> [connector](#output\_connector) | A new Fivetran connector id |
| <a name="output_destination"></a> [destination](#output\_destination) | A new Fivetran destination id |
| <a name="output_user"></a> [user](#output\_user) | A new Fivetran user id |
| <a name="output_group"></a> [public\_group](#output\_group) | A new Fivetran group id |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
