package resources

import (
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWarnAuthFieldsInConfig(t *testing.T) {
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()

	cases := []struct {
		name        string
		service     string
		field       string
		value       attr.Value
		wantWarning bool
	}{
		{
			name:        "auth-only field warns",
			service:     "google_play",
			field:       "authentication_method",
			value:       types.StringValue("FIVETRAN_SERVICE_ACCOUNT"),
			wantWarning: true,
		},
		{
			name:        "unknown auth-only field warns",
			service:     "google_play",
			field:       "authentication_method",
			value:       types.StringUnknown(),
			wantWarning: true,
		},
		{
			name:        "sensitive auth-only field warns",
			service:     "asana",
			field:       "refresh_token",
			value:       types.StringValue("token"),
			wantWarning: true,
		},
		{
			name:    "null auth-only field doesn't warn",
			service: "google_play",
			field:   "authentication_method",
			value:   types.StringNull(),
		},
		{
			name:    "field expected in both config and auth doesn't warn",
			service: "share_point",
			field:   "tenant_id",
			value:   types.StringValue("tenant"),
		},
		{
			name:    "same field for a service that expects it in config doesn't warn",
			service: "adobe_analytics",
			field:   "authentication_method",
			value:   types.StringValue("OAUTH"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.ObjectValueMust(
				map[string]attr.Type{tc.field: types.StringType},
				map[string]attr.Value{tc.field: tc.value},
			)
			var diags diag.Diagnostics
			warnAuthFieldsInConfig(config, types.StringValue(tc.service), path.Root("config"), &diags)
			if diags.HasError() {
				t.Fatalf("expected no errors, got diagnostics: %v", diags)
			}
			if (diags.WarningsCount() > 0) != tc.wantWarning {
				t.Fatalf("expected warning: %v, got diagnostics: %v", tc.wantWarning, diags)
			}
		})
	}
}
