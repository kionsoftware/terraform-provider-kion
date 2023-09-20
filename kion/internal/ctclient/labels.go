package ctclient

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var supportedResourceTypes = []string{"account", "cloud-rule", "funding-source", "ou", "project"}

func BuildAppLabelIDs(c *Client, d *schema.ResourceData) ([]int, error) {
	labelResp := new(LabelListResponse)
	err := c.GET("/v1/app-label", labelResp)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v", err.Error()))
	}

	appLabelIDs := make([]int, 0)
	labels := d.Get("labels").(map[string]interface{})
	for k, v := range labels {
		match, err := func(k string, v string) (int, error) {
			for _, item := range labelResp.Data.Items {
				if k == item.Key && v == item.Value {
					return item.ID, nil
				}
			}
			return -1, errors.New(fmt.Sprintf("A label with the key %s and value %s does not yet exist", k, v))
		}(k, v.(string))

		if err != nil {
			return nil, err
		} else {
			appLabelIDs = append(appLabelIDs, match)
		}
	}

	return appLabelIDs, nil
}

func PutAppLabelIDs(c *Client, appLabelIDs []int, resourceType string, resourceID string) error {
	if IsSupportedResourceType(resourceType) != true {
		return errors.New(fmt.Sprintf("Error: %v", "Unsupported resource type for labels"))
	}

	req := AppLabelIdsCreate{
		LabelIDs: appLabelIDs,
	}

	err := c.PUT(fmt.Sprintf("/v1/%s/%s/label", resourceType, resourceID), req)
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

func ReadResourceLabels(c *Client, resourceType string, resourceID string) (map[string]interface{}, error) {
	if IsSupportedResourceType(resourceType) != true {
		return nil, errors.New(fmt.Sprintf("Error: %v", "Unsupported resource type for labels"))
	}

	labelsResp := new(ResourceLabelsResponse)
	err := c.GET(fmt.Sprintf("/v3/%s/%s/labels", resourceType, resourceID), labelsResp)
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
