package testdata

// Miniature stand-in for the ogen oas_parameters_gen.go, Label ops only.

type GetLabelParams struct{ ID int64 }

type PatchLabelParams struct{ ID int64 }

type DeleteLabelParams struct{ ID int64 }

type GetLabelIndexParams struct {
	Page  OptInt64 `json:",omitempty,omitzero"`
	Count OptInt64 `json:",omitempty,omitzero"`
}

// GetLabelChildParams paginates, but its parent id is required — the data source
// has no value to put there, so the op is unusable as a whole-collection read.
type GetLabelChildParams struct {
	ID    int64
	Page  OptInt64 `json:",omitempty,omitzero"`
	Count OptInt64 `json:",omitempty,omitzero"`
}
