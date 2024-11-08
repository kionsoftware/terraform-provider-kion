package kionclient

// CustomVariableOverrideListResponse for: GET /api/v3/{entity_type}/{entity_id}/custom-variable
type CustomVariableOverrideListResponse struct {
	Data struct {
		Items []struct {
			CustomVariableID   int    `json:"custom_variable_id"`
			CustomVariableType string `json:"custom_variable_type"`
			Inherited          struct {
				EntityID   int         `json:"entity_id"`
				EntityType string      `json:"entity_type"`
				EntityName string      `json:"entity_name"`
				Value      interface{} `json:"value"`
			}
			Override struct {
				Value interface{} `json:"value"`
			}
			Value interface{} `json:"value"`
		} `json:"items"`
	} `json:"data"`
	Status int `json:"status"`
}

// CustomVariableOverrideResponse for: GET /api/v3/{entity_type}/{entity_id}/custom-variable/{cv_id}
type CustomVariableOverrideResponse struct {
	Data struct {
		CustomVariableID   int    `json:"custom_variable_id"`
		CustomVariableType string `json:"custom_variable_type"`
		Inherited          struct {
			EntityID   int         `json:"entity_id"`
			EntityType string      `json:"entity_type"`
			EntityName string      `json:"entity_name"`
			Value      interface{} `json:"value"`
		}
		Override struct {
			Value interface{} `json:"value"`
		}
		Value interface{} `json:"value"`
	} `json:"data"`
	Status int `json:"status"`
}

// CustomVariableOverrideSet for: PUT /api/v3/{entity_type}/{entity_id}/custom-variable/{cv_id}
type CustomVariableOverrideSet struct {
	Value interface{} `json:"value"`
}
