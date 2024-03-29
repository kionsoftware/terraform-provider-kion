package kionclient

import (
	"errors"
	"fmt"
)

var supportedResourceTypes = []string{"account", "cloud-rule", "funding-source", "ou", "project"}

func PutAppLabelIDs(k *Client, labels *[]AssociateLabel, resourceType string, resourceID string) error {
	if IsSupportedResourceType(resourceType) != true {
		return errors.New(fmt.Sprintf("Error: %v", "Unsupported resource type for labels"))
	}

	req := AssociateLabels{
		Labels: labels,
	}

	err := k.PUT(fmt.Sprintf("/v3/%s/%s/labels", resourceType, resourceID), req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %v", err.Error()))
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

func ReadResourceLabels(k *Client, resourceType string, resourceID string) (map[string]interface{}, error) {
	if IsSupportedResourceType(resourceType) != true {
		return nil, errors.New(fmt.Sprintf("Error: %v", "Unsupported resource type for labels"))
	}

	labelsResp := new(AssociatedLabelsResponse)
	err := k.GET(fmt.Sprintf("/v3/%s/%s/labels", resourceType, resourceID), labelsResp)
	if err != nil {
		return nil, err
	}

	labelItems := labelsResp.Data
	labelData := make(map[string]interface{}, 0)
	for _, item := range labelItems {
		labelData[item.Key] = item.Value
	}

	return labelData, nil
}
