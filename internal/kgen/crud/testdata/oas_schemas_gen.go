package testdata

// Miniature stand-in for the ogen oas_schemas_gen.go, Label ops only. Parsed
// as bytes by the crud generator (never compiled — under testdata/).

type CreateLabel struct {
	Color string    `json:"color"`
	ID    OptUint64 `json:"id"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type UpdateLabel struct {
	Color OptString `json:"color"`
	Key   OptString `json:"key"`
	Value OptString `json:"value"`
}

// Label mirrors CreateLabel: required scalars are plain, the server-assigned
// id is Opt-wrapped (matches the real SDK, where flatten uses StringToFramework).
type Label struct {
	Color string    `json:"color"`
	ID    OptUint64 `json:"id"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type LabelResponse struct {
	Data   OptLabel `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*LabelResponse) getLabelRes() {}

type (
	OptString struct{ Value string }
	OptUint64 struct{ Value uint64 }
	OptInt64  struct{ Value int64 }
	OptLabel  struct{ Value Label }
)

type OptNilLabelArray struct {
	Value []Label
	Set   bool
	Null  bool
}

type LabelListPaginatedResponse struct {
	Data   OptLabelListPaginated `json:"data"`
	Status OptInt64              `json:"status"`
}

func (*LabelListPaginatedResponse) getLabelIndexRes() {}

type LabelListPaginated struct {
	Items OptNilLabelArray `json:"items"`
	Total OptInt64         `json:"total"`
}

type OptLabelListPaginated struct {
	Value LabelListPaginated
	Set   bool
}

// LabelFlatListResponse is the second envelope shape the SDK uses: `data` is the
// items slice itself, with no inner pagination struct and no total.
type LabelFlatListResponse struct {
	Data   []Label  `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*LabelFlatListResponse) getLabelFlatRes() {}

type LabelChildListResponse struct {
	Data   []Label  `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*LabelChildListResponse) getLabelChildRes() {}

type Other struct {
	ID OptUint64 `json:"id"`
}

type LabelForeignListResponse struct {
	Data   []Other  `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*LabelForeignListResponse) getLabelForeignRes() {}
