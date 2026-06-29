package framework

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerSchema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func TestProviderSchemaIncludesSkipPlanTimeValidation(t *testing.T) {
	t.Parallel()

	p := &fivetranProvider{metadataCache: &sync.Map{}}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["skip_plan_time_validation"].(providerSchema.BoolAttribute)
	if !ok {
		t.Fatalf("skip_plan_time_validation has type %T, want BoolAttribute", resp.Schema.Attributes["skip_plan_time_validation"])
	}
	if !attr.Optional || attr.Required {
		t.Fatalf("skip_plan_time_validation mode = required:%v optional:%v, want optional only", attr.Required, attr.Optional)
	}
}
