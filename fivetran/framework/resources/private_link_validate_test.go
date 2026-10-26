package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWarnDeprecatedPrivateLinkConfigKeys(t *testing.T) {
	t.Parallel()

	stringMap := func(elements map[string]attr.Value) types.Map {
		m, d := types.MapValue(types.StringType, elements)
		if d.HasError() {
			t.Fatalf("failed to build map: %v", d)
		}
		return m
	}

	cases := []struct {
		name      string
		configMap types.Map
		warnings  int
	}{
		{
			name:      "deprecated key",
			configMap: stringMap(map[string]attr.Value{"private_dns_regions": types.StringValue("x")}),
			warnings:  1,
		},
		{
			name:      "deprecated key with unknown value",
			configMap: stringMap(map[string]attr.Value{"private_dns_regions": types.StringUnknown()}),
			warnings:  1,
		},
		{
			name:      "supported key",
			configMap: stringMap(map[string]attr.Value{"private_dns_records": types.StringValue("a;b")}),
			warnings:  0,
		},
		{
			name:      "null map",
			configMap: types.MapNull(types.StringType),
			warnings:  0,
		},
		{
			name:      "unknown map",
			configMap: types.MapUnknown(types.StringType),
			warnings:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diags diag.Diagnostics
			warnDeprecatedPrivateLinkConfigKeys(tc.configMap, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected errors: %v", diags)
			}
			if diags.WarningsCount() != tc.warnings {
				t.Fatalf("warnings = %d, want %d: %v", diags.WarningsCount(), tc.warnings, diags)
			}
			if tc.warnings > 0 {
				want := path.Root("config_map").AtMapKey("private_dns_regions")
				got, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
				if !ok || !got.Path().Equal(want) {
					t.Fatalf("warning path = %v, want %v", got, want)
				}
			}
		})
	}
}
