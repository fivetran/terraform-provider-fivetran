# Local Build Quickstart

## 1. `~/.terraformrc`

```hcl
provider_installation {
  dev_overrides {
    "fivetran/fivetran" = "/Users/jovan/terraform-provider-dev"
  }
  direct {}
}
```

## 2. Build

```bash
cd /Users/jovan/Fivetran/terraform-provider-fivetran
go build -o /Users/jovan/terraform-provider-dev/terraform-provider-fivetran .
```

## 3. Run

```bash
cd /Users/jovan/terraform-provider-dev/test-workspace
terraform plan
terraform apply
```

No `terraform init` needed — dev overrides skip the registry.

## Iterate

Edit code → rebuild (step 2) → `terraform plan`/`apply` again. Same binary path, no version bump needed.

## Go back to the released provider

Remove or comment out the `dev_overrides` block in `~/.terraformrc`:

```hcl
provider_installation {
  direct {}
}
```

Then re-init so Terraform pulls the real registry version:

```bash
cd /path/to/your/workspace
rm -rf .terraform .terraform.lock.hcl
terraform init
```
