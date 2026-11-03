package testdata

type WidgetCreateRequest struct {
	Name string `json:"name"`
}

type WidgetUpdateRequest struct {
	Name OptString `json:"name"`
}

// Widget is the read payload. It carries secret too (as a redacted value),
// which flatten must never assign over the one captured at create.
type Widget struct {
	ID     OptInt64  `json:"id"`
	Name   OptString `json:"name"`
	Secret OptString `json:"secret"`
}

type WidgetCreated struct {
	ID     int64        `json:"id"`
	Name   string       `json:"name"`
	Secret OptNilString `json:"secret"`
	Count  OptInt64     `json:"count"`
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
	OptNilString struct {
		Value string
		Set   bool
		Null  bool
	}
	OptInt64 struct {
		Value int64
		Set   bool
	}
	OptWidget struct {
		Value Widget
		Set   bool
	}
)
