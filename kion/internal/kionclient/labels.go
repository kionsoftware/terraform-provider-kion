package kionclient

import (
	"fmt"
)

var supportedResourceTypes = []string{"account", "cloud-rule", "funding-source", "ou", "project"}

func PutAppLabelIDs(client *Client, labels *[]AssociateLabel, resourceType string, resourceID string) error {
	if !IsSupportedResourceType(resourceType) {
		return fmt.Errorf("Error: %v", "Unsupported resource type for labels")
	}

	// Ensure labels exist
	err := EnsureLabelsExist(client, labels)
	if err != nil {
		return fmt.Errorf("Error ensuring labels exist: %v", err)
	}

	req := AssociateLabels{
		Labels: labels,
	}

	err = client.PUT(fmt.Sprintf("/v3/%s/%s/labels", resourceType, resourceID), req)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	return nil
}

func IsSupportedResourceType(resourceType string) bool {
	for _, item := range supportedResourceTypes {
		if resourceType == item {
			return true
		}
	}
	return false
}

func ReadResourceLabels(client *Client, resourceType string, resourceID string) (map[string]interface{}, error) {
	if !IsSupportedResourceType(resourceType) {
		return nil, fmt.Errorf("Error: %v", "Unsupported resource type for labels")
	}

	labelsResp := new(AssociatedLabelsResponse)
	err := client.GET(fmt.Sprintf("/v3/%s/%s/labels", resourceType, resourceID), labelsResp)
	if err != nil {
		return nil, err
	}

	labelItems := labelsResp.Data
	labelData := make(map[string]interface{})
	for _, item := range labelItems {
		labelData[item.Key] = item.Value
	}

	return labelData, nil
}

func EnsureLabelsExist(client *Client, labels *[]AssociateLabel) error {
	for _, label := range *labels {
		// Check if the label exists
		exists, _, err := LabelExists(client, label.Key, label.Value)
		if err != nil {
			return err
		}
		if !exists {
			// Create the label with a default color
			newLabel := LabelCreate{
				Key:   label.Key,
				Value: label.Value,
				Color: "#000000", // Default color or handle as needed
			}
			_, err := client.POST("/v3/label", newLabel)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func LabelExists(client *Client, key string, value string) (bool, int, error) {
	resp := new(LabelListResponse)
	err := client.GET("/v3/label", resp)
	if err != nil {
		return false, 0, err
	}
	for _, label := range resp.Data.Items {
		if label.Key == key && label.Value == value {
			return true, label.ID, nil
		}
	}
	return false, 0, nil
}
