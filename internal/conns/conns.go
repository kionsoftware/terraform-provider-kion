// Package conns defines the service package interface and shared connection types.
package conns

import (
	"context"

	"terraform-provider-kion/internal/servicepkg"
)

// ServicePackage is the interface that each service package must implement
// to register its resources and data sources with the provider.
type ServicePackage interface {
	Resources(context.Context) []servicepkg.ServicePackageResource
	DataSources(context.Context) []servicepkg.ServicePackageDataSource
	ServicePackageName() string
}
