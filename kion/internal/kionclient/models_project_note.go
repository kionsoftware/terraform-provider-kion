package kionclient

// ProjectNoteListResponse for: GET /api/v3/project-note
type ProjectNoteListResponse struct {
	Data []struct {
		CreateUserID     uint   `json:"create_user_id"`
		CreateUserName   string `json:"create_user_name"`
		ID               uint   `json:"id"`
		LastUpdateUserID struct {
			Int   uint `json:"Int"`
			Valid bool `json:"Valid"`
		} `json:"last_update_user_id"`
		LastUpdateUserName string `json:"last_update_user_name"`
		Name               string `json:"name"`
		ProjectID          uint   `json:"project_id"`
		Text               string `json:"text"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectNoteResponse for: GET /api/v3/project-note/{id}
type ProjectNoteResponse struct {
	Data struct {
		CreateUserID     uint   `json:"create_user_id"`
		CreateUserName   string `json:"create_user_name"`
		ID               uint   `json:"id"`
		LastUpdateUserID struct {
			Int   uint `json:"Int"`
			Valid bool `json:"Valid"`
		} `json:"last_update_user_id"`
		LastUpdateUserName string `json:"last_update_user_name"`
		Name               string `json:"name"`
		ProjectID          uint   `json:"project_id"`
		Text               string `json:"text"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectNoteCreate for: POST /api/v3/project-note
type ProjectNoteCreate struct {
	CreateUserID uint   `json:"create_user_id"`
	Name         string `json:"name"`
	ProjectID    uint   `json:"project_id"`
	Text         string `json:"text"`
}

// ProjectNoteUpdate for: PATCH /api/v2/project-note/{id}
type ProjectNoteUpdate struct {
	ID           uint   `json:"id"`
	CreateUserID uint   `json:"create_user_id"`
	ProjectID    uint   `json:"project_id"`
	Name         string `json:"name"`
	Text         string `json:"text"`
}
