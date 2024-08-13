package kionclient

// UserListResponse for: GET /api/v3/user
type UserListResponse struct {
	Data []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Enabled  bool   `json:"enabled"`
	} `json:"data"`
	Status int `json:"status"`
}

// UserResponse for: GET /api/v3/user/{id}
type UserResponse struct {
	Data struct {
		User struct {
			ID int `json:"id"`
		} `json:"user"`
	} `json:"data"`
	Status int `json:"status"`
}
