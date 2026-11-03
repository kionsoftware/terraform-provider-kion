package testdata

type ThingCreateRequest struct {
	Description OptString `json:"description"`
	Name        string    `json:"name"`
}

// ThingUpdateRequest wraps the record in an object that carries its id, like
// PermissionSchemeUpdateRequest.app_policy does with an integer one.
type ThingUpdateRequest struct {
	Thing OptThingRef `json:"thing"`
}

type ThingRef struct {
	Description OptString `json:"description"`
	ID          OptString `json:"id"`
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
	OptThingRef struct {
		Value ThingRef
		Set   bool
	}
	OptThingCreated struct {
		Value ThingCreated
		Set   bool
	}
)
