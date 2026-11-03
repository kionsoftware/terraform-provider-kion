package crud

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
)

// enforcementData is a minimal parent_list payload shaped like
// project_enforcement, enough to render the template.
func enforcementData() parentListData {
	return parentListData{
		Pkg: "project_enforcement", Pascal: "ProjectEnforcement", Model: "ProjectEnforcementModel",
		ResConst: "ResNameProjectEnforcement", ResName: "ProjectEnforcement Resource",
		TypeName: "kion_project_enforcement", SDKAlias: "generated",
		IDGo: "Id", ParentIDGo: "ProjectId", ParentIDTF: "project_id",
		ParentParam: "ID", ChildParam: "EnforcementID",
		CreateMethod: "PostProjectEnforcement", CreateBody: "ProjectEnforcementCreate",
		CreateBodyOpt: "OptProjectEnforcementCreate", CreateParams: "PostProjectEnforcementParams",
		ReadMethod: "GetProjectEnforcements", ReadParams: "GetProjectEnforcementsParams",
		ResponseType: "ProjectEnforcementResponse", RecordType: "ProjectEnforcement", RecordIDGo: "ID",
		HasUpdate: true, UpdateMethod: "PatchProjectEnforcements", UpdateBody: "ProjectEnforcementUpdate",
		UpdateBodyOpt: "OptProjectEnforcementUpdate", UpdateParams: "PatchProjectEnforcementsParams",
		UpdateBinds:  []fieldBind{{SDKField: "Enabled", Converter: "flex.OptNilBoolFromFramework", ModelGo: "Enabled"}},
		DeleteMethod: "DeleteProjectEnforcements", DeleteParams: "DeleteProjectEnforcementsParams",
	}
}

func renderParentList(t *testing.T, d parentListData) []byte {
	t.Helper()
	got, err := execGoTemplate("parentlist", parentListTmpl, d, "x.go")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "x.go", got, parser.ParseComments); err != nil {
		t.Fatalf("does not parse: %v\n%s", err, got)
	}
	return got
}

// The update body carries no notification recipients; they sync through the
// enforcement's own add/remove endpoints, addressed by parent and child id.
func TestRenderParentList_assocs(t *testing.T) {
	d := enforcementData()
	d.Assocs = []*assocMembershipBind{{
		AddMethod: "PostProjectEnforcementUsers", RemoveMethod: "DeleteProjectEnforcementUsers",
		Body: "ProjectEnforcementUsers", BodyOpt: "OptProjectEnforcementUsers",
		AddParams: "PostProjectEnforcementUsersParams", RemoveParams: "DeleteProjectEnforcementUsersParams",
		Fields: []assocField{{ModelGo: "UserIds", BodyGo: "UserIds", Coll: "Set"}},
	}}
	got := renderParentList(t, d)
	for _, w := range []string{
		"req.State.Get(ctx, &state)",
		"flex.Uint64SetDiff(ctx, state.UserIds, plan.UserIds, &resp.Diagnostics)",
		"conn.PostProjectEnforcementUsers(ctx, generated.OptProjectEnforcementUsers{",
		"generated.PostProjectEnforcementUsersParams{ID: parentID, EnforcementID: id}",
		"conn.DeleteProjectEnforcementUsers(ctx,",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("missing %q\n%s", w, got)
		}
	}
}

// A field only the update body carries is applied by an update straight after
// the create when configured, rather than dropped.
func TestRenderParentList_createViaUpdate(t *testing.T) {
	d := enforcementData()
	d.CreateViaUpdate = []string{"Enabled"}
	got := renderParentList(t, d)
	for _, w := range []string{
		"!plan.Enabled.IsNull() && !plan.Enabled.IsUnknown()",
		"expandProjectEnforcementUpdate(ctx, plan)",
		"func expandProjectEnforcementUpdate(",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("missing %q\n%s", w, got)
		}
	}
	// Update sends the update once; Create adds a second only when declared. The
	// call stays inline in each so config derivation still sees Update's op.
	call := []byte("conn.PatchProjectEnforcements(ctx, input,")
	if n := bytes.Count(got, call); n != 2 {
		t.Errorf("want 2 update calls with create_via_update, got %d", n)
	}
	if n := bytes.Count(renderParentList(t, enforcementData()), call); n != 1 {
		t.Errorf("want 1 update call without create_via_update, got %d", n)
	}
}
