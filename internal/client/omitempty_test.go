package client

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// overlay.yaml gives every nullable property `omitempty`, then takes it away
// again from the ones that are also REQUIRED, because for those a null has to
// reach the wire. That second list is written by hand, so these tests pin it.

// requiredNullable lists every property in the vendored spec that is both
// required and nullable. Each one needs an `x-omitempty: false` action in
// overlay.yaml.
var requiredNullable = []string{
	"components.schemas.SharedStorageQuota.configured_gb",
	"components.schemas.SharedStorageQuota.observed_quota_bytes",
	"components.schemas.SharedStorageQuota.observed_used_bytes",
	"paths./api/v1/admin/organizations/{orgID}/storage-quota.put.requestBody.content.application/json.schema.quota_gb",
}

func isNullable(prop map[string]any) bool {
	if types, ok := prop["type"].([]any); ok {
		for _, t := range types {
			if t == "null" {
				return true
			}
		}
	}
	for _, key := range []string{"oneOf", "anyOf"} {
		alts, _ := prop[key].([]any)
		for _, alt := range alts {
			if m, ok := alt.(map[string]any); ok && m["type"] == "null" {
				return true
			}
		}
	}
	return false
}

func collectRequiredNullable(node any, path string, out *[]string) {
	switch n := node.(type) {
	case map[string]any:
		if props, ok := n["properties"].(map[string]any); ok {
			required, _ := n["required"].([]any)
			for _, name := range required {
				if prop, ok := props[name.(string)].(map[string]any); ok && isNullable(prop) {
					*out = append(*out, path+"."+name.(string))
				}
			}
		}
		for key, child := range n {
			collectRequiredNullable(child, strings.TrimPrefix(path+"."+key, "."), out)
		}
	case []any:
		for _, child := range n {
			collectRequiredNullable(child, path, out)
		}
	}
}

func TestContract_RequiredNullableFieldsAreListedInTheOverlay(t *testing.T) {
	b, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]any
	if err := yaml.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	var got []string
	collectRequiredNullable(spec, "", &got)
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(requiredNullable, "\n") {
		t.Fatalf("required+nullable properties in api/openapi.yaml changed.\n got:\n  %s\nwant:\n  %s\n"+
			"Add or remove the matching `x-omitempty: false` action in internal/client/overlay.yaml, "+
			"run `go generate ./...`, then update requiredNullable in this file.",
			strings.Join(got, "\n  "), strings.Join(requiredNullable, "\n  "))
	}
}

// A required nullable field left unset must be sent as an explicit null: a
// null quota_gb is how an org is reverted to the platform default.
func TestContract_RequiredNullableFieldsKeepNull(t *testing.T) {
	for _, tc := range []struct {
		value any
		keys  []string
	}{
		{AdminSetSharedStorageQuotaJSONBody{}, []string{"quota_gb"}},
		{SharedStorageQuota{}, []string{"configured_gb", "observed_quota_bytes", "observed_used_bytes"}},
	} {
		b, err := json.Marshal(tc.value)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatal(err)
		}
		for _, key := range tc.keys {
			if v, present := got[key]; !present || v != nil {
				t.Errorf("%T: %q must marshal as null when unset, got %s", tc.value, key, b)
			}
		}
	}
}

// An optional nullable field left unset must stay out of the request. To the
// API an explicit null means "clear this value", so sending one for every
// field the provider did not set would wipe fields Terraform does not manage.
func TestContract_OptionalNullableFieldsAreOmitted(t *testing.T) {
	for _, value := range []any{UpdateOrganizationRequest{}, CreateOrganizationRequest{}} {
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "null") {
			t.Errorf("%T marshals unset optional fields as null: %s", value, b)
		}
	}
}
