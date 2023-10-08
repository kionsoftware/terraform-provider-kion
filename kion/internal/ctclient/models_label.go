package ctclient

type Label struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Color string `json:"color"`
}

// LabelListResponse for: GET /api/v3/label
type LabelListResponse struct {
	Data struct {
		Items []Label `json:"items"`
		Total int     `json:"total"`
	} `json:"data"`
	Status int `json:"status"`
}

// LabelResponse for: GET /api/v3/label/{id}
type LabelResponse struct {
	Data   Label `json:"data"`
	Status int   `json:"status"`
}

// LabelCreate for: POST /api/v3/label
type LabelCreate struct {
	Color string `json:"color"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LabelUpdatable for: PATCH /api/v3/label/{id}
type LabelUpdatable struct {
	Color string `json:"color"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
