package crud

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// membershipOwners declares a resource's owner association, synced on Update via
// paired add/remove endpoints (the main update body does not carry owners). The
// two model attributes are combined into one {owner_user_ids, owner_user_group_ids}
// struct-body call per direction.
type membershipOwners struct {
	UserField  string `yaml:"user_field"`  // model attr for owner user ids
	GroupField string `yaml:"group_field"` // model attr for owner user-group ids
	Add        string `yaml:"add"`         // SDK add method, e.g. PostAzurePolicyOwners
	Remove     string `yaml:"remove"`      // SDK remove method, e.g. DeleteAzurePolicyOwners
	Body       string `yaml:"body"`        // owners body struct, e.g. AzurePolicyDefinitionOwners
	Ptr        bool   `yaml:"ptr"`         // the endpoint takes *Body (pointer) rather than an Opt<Body> wrapper
}

// membershipAssociations declares a bulk association endpoint that syncs many
// id-list attributes at once (e.g. cloud_rule: cft_ids, iam_policy_ids, … via
// PostCloudRuleAssociations/DeleteCloudRuleAssociations with a CloudRuleAssociations
// body). Each field is diffed state-vs-plan; the added deltas are POSTed and the
// removed deltas DELETEd in one call each. The create body already sets initial
// values, so this only runs on Update.
type membershipAssociations struct {
	Add    string   `yaml:"add"`    // SDK add method, e.g. PostCloudRuleAssociations
	Remove string   `yaml:"remove"` // SDK remove method
	Body   string   `yaml:"body"`   // associations body struct, e.g. CloudRuleAssociations
	Ptr    bool     `yaml:"ptr"`    // endpoint takes *Body rather than Opt<Body>
	Fields []string `yaml:"fields"` // model attrs (each an int id list; tfsdk name == body json name)
	// FieldJSON maps a model attribute to a body field whose json name differs,
	// for the cases where the two spellings diverge (the provider's
	// car_restricted_user_group_ids is car_restricted_ugroup_ids on the wire).
	FieldJSON map[string]string `yaml:"field_json"`
	// DetachBeforeDelete makes Delete remove these associations first.
	//
	// Kion refuses to delete a record something else still points at --
	// "this cloud rule is in use, and cannot be deleted" -- and Terraform has no
	// way to know that: the associations are attributes of THIS resource, not
	// separate resources it can order a destroy around. So `terraform destroy`
	// failed and left the record behind, needing a manual API call to unpick.
	//
	// Observed in the live migration harness, where destroying a cloud rule
	// attached to an OU, a project and a compliance standard failed outright.
	DetachBeforeDelete bool `yaml:"detach_before_delete"`
}

// membershipSlice declares a member id-list synced via []int64 add/remove
// endpoints (e.g. user_group users: UpdateUGroupUsers(ctx, []int64, params) /
// RemoveUGroupUsers). Diffed state-vs-plan; added ids POSTed, removed DELETEd.
type membershipSlice struct {
	Field  string `yaml:"field"`  // model attr (int id list)
	Add    string `yaml:"add"`    // SDK add method taking []int64
	Remove string `yaml:"remove"` // SDK remove method taking []int64
}

// membershipConfig is one resource's membership declarations. Associations is a
// list so a resource can sync several struct-bodied endpoints (e.g. user_group's
// owners + viewers).
// membershipLabels declares the per-resource label sub-resource. Kion carries
// labels at GET/PUT /v3/<type>/{id}/labels, never in the resource's own request
// body, so a resource exposing `labels` must sync them separately -- without
// which the attribute is accepted and silently discarded.
type membershipLabels struct {
	Attr    string `yaml:"attr"`    // model attribute, always "labels" today
	Get     string `yaml:"get"`     // SDK read op, e.g. GetOULabels
	Put     string `yaml:"put"`     // SDK write op, e.g. PutOULabels
	Params  string `yaml:"params"`  // params field holding the parent id
	Element string `yaml:"element"` // per-resource record type the GET returns
}

type membershipConfig struct {
	Labels       *membershipLabels        `yaml:"labels"`
	Owners       *membershipOwners        `yaml:"owners"`
	Associations []membershipAssociations `yaml:"associations"`
	SliceMembers []membershipSlice        `yaml:"slice_members"`
}

// loadMemberships reads codegen/memberships.yaml. Missing file → empty.
func loadMemberships(path string) (map[string]membershipConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]membershipConfig{}, nil
		}
		return nil, fmt.Errorf("reading memberships %s: %w", path, err)
	}
	var f struct {
		Resources map[string]membershipConfig `yaml:"resources"`
	}
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parsing memberships %s: %w", path, err)
	}
	return f.Resources, nil
}

// ownerMembershipBind is the resolved template payload for an owner sync.
type ownerMembershipBind struct {
	UserFieldGo  string // model GoName, "OwnerUsers"
	GroupFieldGo string // "OwnerUserGroups"
	UserColl     string // "List" | "Set" (flex diff variant for the user field)
	GroupColl    string // "List" | "Set"
	AddMethod    string // "PostAzurePolicyOwners"
	RemoveMethod string // "DeleteAzurePolicyOwners"
	Body         string // "AzurePolicyDefinitionOwners"
	BodyOpt      string // "OptAzurePolicyDefinitionOwners"
	Ptr          bool   // endpoint takes *Body rather than Opt<Body>
	AddParams    string // "PostAzurePolicyOwnersParams"
	RemoveParams string // "DeleteAzurePolicyOwnersParams"
}

// assocField is one id-list attribute synced through the bulk endpoint.
type assocField struct {
	ModelGo string // "CftIds"
	BodyGo  string // SDK body struct field, "CftIds"
	Coll    string // "List" | "Set"
}

// assocMembershipBind is the resolved payload for a bulk association sync.
type assocMembershipBind struct {
	AddMethod    string
	RemoveMethod string
	Body         string
	BodyOpt      string
	Ptr          bool
	AddParams    string
	RemoveParams string
	Fields       []assocField
	// DetachBeforeDelete: Delete removes these associations before deleting.
	DetachBeforeDelete bool
}

// resolveAssocMembership maps a membershipAssociations config onto the model +
// SDK body struct (via idx).
func resolveAssocMembership(ma membershipAssociations, byTF map[string]ModelField, idx sdkIndex) (*assocMembershipBind, error) {
	if ma.Add == "" || ma.Remove == "" || ma.Body == "" || len(ma.Fields) == 0 {
		return nil, fmt.Errorf("associations requires add, remove, body, and fields")
	}
	bodyStruct, ok := idx.structs[ma.Body]
	if !ok {
		return nil, fmt.Errorf("associations body %q not found in SDK", ma.Body)
	}
	bodyByJSON := map[string]Field{}
	for _, bf := range bodyStruct.Fields {
		bodyByJSON[bf.JSONName] = bf
	}
	var fields []assocField
	for _, tf := range ma.Fields {
		mf, ok := byTF[tf]
		if !ok {
			return nil, fmt.Errorf("associations field %q not in model", tf)
		}
		coll := collKind(mf.Type)
		if coll == "" {
			return nil, fmt.Errorf("associations field %q must be list/set, got %q", tf, mf.Type)
		}
		jsonName := tf
		if alt, renamed := ma.FieldJSON[tf]; renamed {
			jsonName = alt
		}
		bf, ok := bodyByJSON[jsonName]
		if !ok {
			return nil, fmt.Errorf("associations field %q (json %q) not in body %q", tf, jsonName, ma.Body)
		}
		fields = append(fields, assocField{ModelGo: mf.GoName, BodyGo: bf.GoName, Coll: coll})
	}
	return &assocMembershipBind{
		AddMethod: ma.Add, RemoveMethod: ma.Remove, Body: ma.Body, BodyOpt: "Opt" + ma.Body, Ptr: ma.Ptr,
		AddParams: ma.Add + "Params", RemoveParams: ma.Remove + "Params", Fields: fields,
		DetachBeforeDelete: ma.DetachBeforeDelete,
	}, nil
}

// sliceMemberBind is the resolved payload for a []int64 member sync.
type sliceMemberBind struct {
	ModelGo      string // "UserIds"
	Coll         string // "List" | "Set"
	AddMethod    string // "UpdateUGroupUsers"
	RemoveMethod string // "RemoveUGroupUsers"
	AddParams    string
	RemoveParams string
	Var          string // "userIds", local diff var base
}

// resolveSliceMember maps a membershipSlice onto the model.
func resolveSliceMember(ms membershipSlice, byTF map[string]ModelField) (*sliceMemberBind, error) {
	if ms.Field == "" || ms.Add == "" || ms.Remove == "" {
		return nil, fmt.Errorf("slice member requires field, add, remove")
	}
	mf, ok := byTF[ms.Field]
	if !ok {
		return nil, fmt.Errorf("slice member field %q not in model", ms.Field)
	}
	coll := collKind(mf.Type)
	if coll == "" {
		return nil, fmt.Errorf("slice member field %q must be list/set, got %q", ms.Field, mf.Type)
	}
	return &sliceMemberBind{
		ModelGo: mf.GoName, Coll: coll, AddMethod: ms.Add, RemoveMethod: ms.Remove,
		AddParams: ms.Add + "Params", RemoveParams: ms.Remove + "Params", Var: lowerFirst(mf.GoName),
	}, nil
}

// resolveOwnerMembership maps a membershipOwners config onto the model + SDK.
func resolveOwnerMembership(mo membershipOwners, byTF map[string]ModelField) (*ownerMembershipBind, error) {
	uf, ok := byTF[mo.UserField]
	if !ok {
		return nil, fmt.Errorf("owners user_field %q not in model", mo.UserField)
	}
	gf, ok := byTF[mo.GroupField]
	if !ok {
		return nil, fmt.Errorf("owners group_field %q not in model", mo.GroupField)
	}
	uc, gc := collKind(uf.Type), collKind(gf.Type)
	if uc == "" || gc == "" {
		return nil, fmt.Errorf("owners fields %q/%q must be list/set, got %q/%q", mo.UserField, mo.GroupField, uf.Type, gf.Type)
	}
	if mo.Add == "" || mo.Remove == "" || mo.Body == "" {
		return nil, fmt.Errorf("owners requires add, remove, and body")
	}
	return &ownerMembershipBind{
		UserFieldGo: uf.GoName, GroupFieldGo: gf.GoName, UserColl: uc, GroupColl: gc,
		AddMethod: mo.Add, RemoveMethod: mo.Remove, Body: mo.Body, BodyOpt: "Opt" + mo.Body, Ptr: mo.Ptr,
		AddParams: mo.Add + "Params", RemoveParams: mo.Remove + "Params",
	}, nil
}
