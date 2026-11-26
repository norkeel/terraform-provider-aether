package client

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// overlay.yaml gives every nullable property `omitempty`, then takes it away
// again from the ones that are also REQUIRED, because for those a null has to
// reach the wire. Both halves are written by hand - the first as a JSONPath
// that only knows some spellings of "nullable", the second as a list of
// properties - so these tests pin the generated client that results.

// requiredNullable maps every property in the vendored spec that is both
// required and nullable to the generated field it becomes, written
// "GoType.json_name". Each one needs an `x-omitempty: false` action in
// overlay.yaml.
var requiredNullable = map[string]string{
	"components.schemas.SharedStorageQuota.configured_gb":        "SharedStorageQuota.configured_gb",
	"components.schemas.SharedStorageQuota.observed_quota_bytes": "SharedStorageQuota.observed_quota_bytes",
	"components.schemas.SharedStorageQuota.observed_used_bytes":  "SharedStorageQuota.observed_used_bytes",

	"paths./api/v1/admin/organizations/{orgID}/storage-quota.put.requestBody.content.application/json.schema.quota_gb": "AdminSetSharedStorageQuotaJSONBody.quota_gb",
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

func TestContract_RequiredNullableFieldsAreKnown(t *testing.T) {
	b, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]any
	if err := yaml.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	var got, want []string
	collectRequiredNullable(spec, "", &got)
	sort.Strings(got)
	for path := range requiredNullable {
		want = append(want, path)
	}
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("required+nullable properties in api/openapi.yaml changed.\n got:\n  %s\nwant:\n  %s\n"+
			"Add or remove the matching `x-omitempty: false` action in internal/client/overlay.yaml, "+
			"run `go generate ./...`, then update requiredNullable in this file.",
			strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}

// pointerFieldsWithoutOmitempty returns every struct field of the generated
// client that is a pointer and whose JSON tag has no `omitempty`, written
// "GoType.json_name" after the top-level type it is declared in. Left nil,
// such a field is sent as an explicit null.
func pointerFieldsWithoutOmitempty(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "client.gen.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			ast.Inspect(typeSpec.Type, func(n ast.Node) bool {
				field, ok := n.(*ast.Field)
				if !ok || field.Tag == nil {
					return true
				}
				if _, isPointer := field.Type.(*ast.StarExpr); !isPointer {
					return true
				}
				tag, err := strconv.Unquote(field.Tag.Value)
				if err != nil {
					t.Fatalf("%s: struct tag %s: %v", typeSpec.Name.Name, field.Tag.Value, err)
				}
				name, options, _ := strings.Cut(reflect.StructTag(tag).Get("json"), ",")
				if name == "" || name == "-" || strings.Contains(","+options+",", ",omitempty,") {
					return true
				}
				seen[typeSpec.Name.Name+"."+name] = true
				return true
			})
		}
	}
	fields := make([]string, 0, len(seen))
	for field := range seen {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

// The whole generated client, not a sample of it: the only pointer fields that
// may lack `omitempty` are the required nullable ones. This is what catches a
// nullable property the first overlay action does not recognise (oapi-codegen
// also follows `$ref` and `allOf` to find a null; the JSONPath cannot), and a
// required nullable property with no `x-omitempty: false` action.
func TestContract_OnlyRequiredNullableFieldsLackOmitempty(t *testing.T) {
	var want []string
	for _, field := range requiredNullable {
		want = append(want, field)
	}
	sort.Strings(want)
	got := pointerFieldsWithoutOmitempty(t)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("pointer fields without `omitempty` in client.gen.go:\n got:\n  %s\nwant:\n  %s\n"+
			"A field only in `got` is an optional nullable property that would be sent as an explicit null "+
			"whenever it is unset: add an `x-omitempty: true` action for it to internal/client/overlay.yaml. "+
			"A field only in `want` is a required nullable property whose null would be dropped: "+
			"add its `x-omitempty: false` action there. Then run `go generate ./...`.",
			strings.Join(got, "\n  "), strings.Join(want, "\n  "))
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
