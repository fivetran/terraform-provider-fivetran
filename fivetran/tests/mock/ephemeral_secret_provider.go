package mock

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralSchema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactoriesWithEphemeralSecret adds the "testsecret" provider to ProtoV6ProviderFactories.
// Its only ephemeral resource returns the configured value as an ephemeral one, for tests of
// write-only attributes:
//
//	ephemeral "testsecret_value" "pw" {
//	  value = "secret"
//	}
//
// ephemeral.testsecret_value.pw.secret is then an ephemeral "secret".
var ProtoV6ProviderFactoriesWithEphemeralSecret = map[string]func() (tfprotov6.ProviderServer, error){
	"fivetran-provider": ProtoV6ProviderFactories["fivetran-provider"],
	"testsecret": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(&ephemeralSecretProvider{})(), nil
	},
}

type ephemeralSecretProvider struct{}

var _ provider.ProviderWithEphemeralResources = &ephemeralSecretProvider{}

func (p *ephemeralSecretProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "testsecret"
}

func (p *ephemeralSecretProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
}

func (p *ephemeralSecretProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *ephemeralSecretProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *ephemeralSecretProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *ephemeralSecretProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		func() ephemeral.EphemeralResource { return &ephemeralSecretValue{} },
	}
}

type ephemeralSecretValue struct{}

type ephemeralSecretValueModel struct {
	Value  types.String `tfsdk:"value"`
	Secret types.String `tfsdk:"secret"`
}

func (e *ephemeralSecretValue) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_value"
}

func (e *ephemeralSecretValue) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = ephemeralSchema.Schema{
		Attributes: map[string]ephemeralSchema.Attribute{
			"value": ephemeralSchema.StringAttribute{
				Required: true,
			},
			"secret": ephemeralSchema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *ephemeralSecretValue) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ephemeralSecretValueModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Secret = data.Value
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
