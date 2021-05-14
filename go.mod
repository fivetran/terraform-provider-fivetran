module github.com/fivetran/terraform-provider-fivetran

replace github.com/fivetran/go-fivetran => /Users/felipen.neuwald/git/go-fivetran

require (
	github.com/fivetran/go-fivetran v0.0.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
)

go 1.16
