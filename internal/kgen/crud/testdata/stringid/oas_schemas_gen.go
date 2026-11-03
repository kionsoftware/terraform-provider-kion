package testdata

type ThingCreateRequest struct {
	Description OptString `json:"description"`
	Name        string    `json:"name"`
}

type ThingUpdateRequest struct {
	Description OptString `json:"description"`
	Name        OptString `json:"name"`
}

type Thing struct {
	Description OptString `json:"description"`
	ID          OptString `json:"id"`
	Name        OptString `json:"name"`
}

// ThingCreated is the create payload; like OAuthClientCreated it carries a
// field the read never returns.
type ThingCreated struct {
	Description OptString `json:"description"`
	ID          OptString `json:"id"`
	Name        OptString `json:"name"`
	Secret      OptString `json:"secret"`
}

type ThingResponse struct {
	Data   OptThing `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*ThingResponse) getThingRes() {}

type ThingCreatedResponse struct {
	Data   OptThingCreated `json:"data"`
	Status OptInt64        `json:"status"`
}

func (*ThingCreatedResponse) postThingRes() {}

type ThingListResponse struct {
	Data   []Thing  `json:"data"`
	Status OptInt64 `json:"status"`
}

func (*ThingListResponse) getThingsRes() {}

type (
	OptString struct {
		Value string
		Set   bool
	}
	OptInt64 struct {
		Value int64
		Set   bool
	}
	OptThing struct {
		Value Thing
		Set   bool
	}
	OptThingCreated struct {
		Value ThingCreated
		Set   bool
	}
)
