package kion

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

func TestAccDataSourceOU(t *testing.T) {
	// Declare a resource to create an OU
	resourceDeclaration := testAccOUGenerateResourceDeclaration(&accTestOU)

	// Declare a data source to retrieve a list of all OUs
	dataSourceDeclarationAll := hc.TestAccOUGenerateDataSourceDeclarationAll(resourceTypeOU, dataSourceLocalNameOU)

	// Declare a data source to retrieve an OU that matches the name filter.
	dataSourceDeclarationFilter := hc.TestAccOUGenerateDataSourceDeclarationFilter(resourceTypeOU, dataSourceLocalNameOU, accTestOU.Name)

	// TestStep: List all OUs
	listOUs := resource.TestStep{
		Config: resourceDeclaration + "\n" + dataSourceDeclarationAll,
		Check: resource.TestCheckResourceAttrSet(
			fmt.Sprintf("data.%v.ous", resourceTypeOU), "list.#",
		),
	}

	// TestStep: Filter all OUs for the OU created by `resourceDeclaration` above.
	filterOUs := resource.TestStep{
		Config: resourceDeclaration + "\n" + dataSourceDeclarationFilter,
		Check: resource.TestCheckResourceAttr(
			fmt.Sprintf("data.%v.ous", resourceTypeOU), "filter.0.values.0", accTestOU.Name,
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOUCheckResourceDestroy,
		Steps: []resource.TestStep{
			listOUs,
			filterOUs,
		},
	})
}
