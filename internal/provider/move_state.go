package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ProviderName is the provider's type name: the prefix of every resource and
// data source type, and the last segment of its registry address.
const ProviderName = "norkeel"

// legacyTypePrefix is the type-name prefix the provider used before the rename
// from Aether to Norkeel (v0.1.x, address plane-shift/aether).
const legacyTypePrefix = "aether_"

// isLegacyProviderAddress reports whether addr is the pre-rename provider,
// on either registry: registry.terraform.io/plane-shift/aether or
// registry.opentofu.org/plane-shift/aether.
func isLegacyProviderAddress(addr string) bool {
	return strings.HasSuffix(addr, "/plane-shift/aether")
}

// moveStateFromLegacy lets a `moved` block carry state from the pre-rename
// `aether_<suffix>` type to this resource, where `norkeel_<suffix>` is the type
// name r's Metadata reports. The schemas did not change in the
// rename, so the old state is decoded with this resource's own schema and
// copied across as is: secrets returned only on create (robot accounts,
// webhooks) are kept, and nothing is replaced or rotated.
//
//	moved {
//	  from = aether_cluster.prod
//	  to   = norkeel_cluster.prod
//	}
//
// Needs Terraform 1.8+ (cross-type moves). A request from any other provider or
// type is left unhandled, so the framework reports it as unsupported.
func moveStateFromLegacy(ctx context.Context, r resource.Resource) []resource.StateMover {
	var mr resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: ProviderName}, &mr)
	legacyType := legacyTypePrefix + strings.TrimPrefix(mr.TypeName, ProviderName+"_")
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	if sr.Diagnostics.HasError() {
		return nil
	}
	src := sr.Schema
	return []resource.StateMover{{
		SourceSchema: &src,
		StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if req.SourceTypeName != legacyType || !isLegacyProviderAddress(req.SourceProviderAddress) {
				return
			}
			// The framework leaves SourceState nil when the legacy state does
			// not decode with this schema. Say so: returning silently would
			// surface as a generic "does not include support" error.
			if req.SourceState == nil {
				resp.Diagnostics.AddError("Cannot move "+legacyType+" state",
					"The "+legacyType+" state could not be decoded with the "+mr.TypeName+" schema, so it was not moved. "+
						"Remove the old address from state and import the resource under "+mr.TypeName+" instead.")
				return
			}
			resp.TargetState.Raw = req.SourceState.Raw
			resp.TargetPrivate = req.SourcePrivate
		},
	}}
}

// Every resource accepts a move from its pre-rename type.
var (
	_ resource.ResourceWithMoveState = &billingBudgetResource{}
	_ resource.ResourceWithMoveState = &clusterResource{}
	_ resource.ResourceWithMoveState = &clusterNetworkResource{}
	_ resource.ResourceWithMoveState = &nodePoolResource{}
	_ resource.ResourceWithMoveState = &organizationResource{}
	_ resource.ResourceWithMoveState = &organizationMemberResource{}
	_ resource.ResourceWithMoveState = &registryResource{}
	_ resource.ResourceWithMoveState = &cleanupPolicyResource{}
	_ resource.ResourceWithMoveState = &robotAccountResource{}
	_ resource.ResourceWithMoveState = &webhookResource{}
)
