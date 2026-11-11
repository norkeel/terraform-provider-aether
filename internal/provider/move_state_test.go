package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// legacyState builds a state for r whose only set attribute is id, decoded
// with r's own schema, as the framework does when SourceSchema is set.
func legacyState(t *testing.T, r resource.Resource) *tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	objType := sr.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for name, at := range objType.AttributeTypes {
		if name == "id" {
			vals[name] = tftypes.NewValue(at, "11111111-2222-3333-4444-555555555555")
			continue
		}
		vals[name] = tftypes.NewValue(at, nil)
	}
	return &tfsdk.State{Schema: sr.Schema, Raw: tftypes.NewValue(objType, vals)}
}

func runMove(t *testing.T, r resource.ResourceWithMoveState, addr, typeName string) *resource.MoveStateResponse {
	t.Helper()
	ctx := context.Background()
	movers := r.MoveState(ctx)
	if len(movers) != 1 || movers[0].SourceSchema == nil {
		t.Fatalf("want one mover with a source schema, got %d", len(movers))
	}
	src := legacyState(t, r)
	resp := &resource.MoveStateResponse{TargetState: tfsdk.State{Schema: src.Schema, Raw: tftypes.NewValue(src.Raw.Type(), nil)}}
	movers[0].StateMover(ctx, resource.MoveStateRequest{
		SourceProviderAddress: addr,
		SourceTypeName:        typeName,
		SourceState:           src,
	}, resp)
	return resp
}

func TestMoveStateFromLegacyTypes(t *testing.T) {
	cases := []struct {
		r      resource.ResourceWithMoveState
		suffix string
	}{
		{&billingBudgetResource{}, "billing_budget"},
		{&clusterResource{}, "cluster"},
		{&clusterNetworkResource{}, "cluster_network"},
		{&nodePoolResource{}, "node_pool"},
		{&organizationResource{}, "organization"},
		{&organizationMemberResource{}, "organization_member"},
		{&registryResource{}, "registry"},
		{&cleanupPolicyResource{}, "registry_cleanup_policy"},
		{&robotAccountResource{}, "registry_robot_account"},
		{&webhookResource{}, "webhook"},
	}
	for _, c := range cases {
		for _, addr := range []string{"registry.terraform.io/plane-shift/aether", "registry.opentofu.org/plane-shift/aether"} {
			resp := runMove(t, c.r, addr, "aether_"+c.suffix)
			if resp.Diagnostics.HasError() {
				t.Fatalf("%s from %s: %v", c.suffix, addr, resp.Diagnostics)
			}
			if resp.TargetState.Raw.IsNull() {
				t.Fatalf("%s from %s: target state not set", c.suffix, addr)
			}
			var got map[string]tftypes.Value
			if err := resp.TargetState.Raw.As(&got); err != nil {
				t.Fatal(err)
			}
			var id string
			if err := got["id"].As(&id); err != nil || id != "11111111-2222-3333-4444-555555555555" {
				t.Fatalf("%s: id = %q (%v), want the moved id", c.suffix, id, err)
			}
		}
	}
}

func TestMoveStateIgnoresOtherSources(t *testing.T) {
	for _, tc := range []struct{ addr, typeName string }{
		{"registry.terraform.io/hashicorp/aws", "aether_cluster"},
		{"registry.terraform.io/plane-shift/aether", "aether_node_pool"},
		{"registry.terraform.io/someone/plane-shift-aether", "aether_cluster"},
	} {
		resp := runMove(t, &clusterResource{}, tc.addr, tc.typeName)
		if !resp.TargetState.Raw.IsNull() {
			t.Fatalf("%s %s: moved state from a source it should not accept", tc.addr, tc.typeName)
		}
	}
}
