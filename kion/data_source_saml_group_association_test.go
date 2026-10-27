package kion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func TestDataSourceSamlGroupAssociationRead(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": 200, "data": [
			{"id": 29, "idms_saml_id": 11, "idms_id": 25, "user_group_id": 168,
			 "assertion_name": "memberOf", "assertion_regex": "^admins$", "should_update_on_login": true},
			{"id": 30, "idms_saml_id": 11, "idms_id": 25, "user_group_id": 169,
			 "assertion_name": "memberOf", "assertion_regex": "^devs$", "should_update_on_login": false}
		]}`))
	}))
	defer server.Close()

	client := hc.NewClient(server.URL, "test-key", "api", false)
	d := schema.TestResourceDataRaw(t, dataSourceSamlGroupAssociation().Schema, map[string]interface{}{
		"idms_id": 25,
		"filter": []interface{}{
			map[string]interface{}{"name": "user_group_id", "values": []interface{}{"168"}},
		},
	})

	diags := dataSourceSamlGroupAssociationRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// The IDMS ID must be substituted into the path, not sent as a literal "{id}".
	if want := "/api/v3/idms/25/group-association"; gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}

	list := d.Get("list").([]interface{})
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1 after filtering on user_group_id", len(list))
	}
	item := list[0].(map[string]interface{})
	if item["id"] != 29 || item["assertion_regex"] != "^admins$" || item["should_update_on_login"] != true {
		t.Errorf("unexpected item: %v", item)
	}
}
