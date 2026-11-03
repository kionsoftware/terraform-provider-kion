package testdata

// Scopes is an OptNil array on both write bodies: unset means "keep the
// server's value", an empty list means "none".
type WidgetCreateRequest struct {
	Name   string            `json:"name"`
	Scopes OptNilStringArray `json:"scopes"`
}

type WidgetUpdateRequest struct {
	Name   OptString         `json:"name"`
	Scopes OptNilStringArray `json:"scopes"`
}

type Widget struct {
	ID     OptInt64          `json:"id"`
	Name   OptString         `json:"name"`
	Scopes OptNilStringArray `json:"scopes"`
}

type WidgetCreated struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type WidgetResponse struct {
	Data   OptWidget `json:"data"`
	Status OptInt64  `json:"status"`
}

func (*WidgetResponse) getWidgetRes() {}

type WidgetCreatedResponse struct {
	Data   *WidgetCreated `json:"data"`
	Status OptInt64       `json:"status"`
}

func (*WidgetCreatedResponse) postWidgetRes() {}

type (
	OptString struct {
		Value string
		Set   bool
	}
	OptInt64 struct {
		Value int64
		Set   bool
	}
	OptNilStringArray struct {
		Value []string
		Set   bool
		Null  bool
	}
	OptWidget struct {
		Value Widget
		Set   bool
	}
)
