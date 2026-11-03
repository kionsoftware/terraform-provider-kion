// Package servicepkg defines shared types for the Terraform provider service packages.
package servicepkg

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ServicePackageResource describes a single resource within a service package.
type ServicePackageResource struct {
	Factory  func() resource.Resource
	TypeName string
}

// ServicePackageDataSource describes a single data source within a service package.
type ServicePackageDataSource struct {
	Factory  func() datasource.DataSource
	TypeName string
}
