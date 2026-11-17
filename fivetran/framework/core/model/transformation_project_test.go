package model

import (
	"context"
	"testing"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTransformationProjectCustomPatchVersionResponse(t *testing.T) {
	for _, test := range []struct {
		name   string
		config map[string]interface{}
		want   types.Bool
	}{
		{name: "missing", want: types.BoolNull()},
		{name: "null", config: map[string]interface{}{"use_latest_patch_version": nil}, want: types.BoolNull()},
		{name: "pinned", config: map[string]interface{}{"use_latest_patch_version": false}, want: types.BoolValue(false)},
		{name: "latest", config: map[string]interface{}{"use_latest_patch_version": true}, want: types.BoolValue(true)},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := transformations.TransformationProjectCustomMergedResponse{}
			response.Data.ProjectConfigCustom = test.config
			resource := TransformationResourceProject{}
			datasource := TransformationDatasourceProject{}

			if diags := resource.ReadFromResponse(context.Background(), response); diags.HasError() {
				t.Fatal(diags)
			}
			if diags := datasource.ReadFromResponse(context.Background(), response); diags.HasError() {
				t.Fatal(diags)
			}
			for name, config := range map[string]types.Object{
				"resource": resource.ProjectConfig, "datasource": datasource.ProjectConfig,
			} {
				if got := config.Attributes()["use_latest_patch_version"]; !test.want.Equal(got) {
					t.Errorf("%s: got %v, want %v", name, got, test.want)
				}
			}
		})
	}
}

func TestTransformationProjectCustomPatchVersionResponseInvalid(t *testing.T) {
	response := transformations.TransformationProjectCustomMergedResponse{}
	response.Data.ProjectConfigCustom = map[string]interface{}{"use_latest_patch_version": "false"}
	resource := TransformationResourceProject{}
	datasource := TransformationDatasourceProject{}
	if diags := resource.ReadFromResponse(context.Background(), response); !diags.HasError() {
		t.Fatal("expected resource diagnostic for a non-boolean patch version setting")
	}
	if diags := datasource.ReadFromResponse(context.Background(), response); !diags.HasError() {
		t.Fatal("expected datasource diagnostic for a non-boolean patch version setting")
	}
}
