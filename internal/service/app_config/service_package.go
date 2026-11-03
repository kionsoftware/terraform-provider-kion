package app_config

import (
	"context"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/servicepkg"
)

var _ conns.ServicePackage = &servicePackage{}

type servicePackage struct{}

// NewServicePackage returns the service package registration.
func NewServicePackage() conns.ServicePackage {
	return &servicePackage{}
}

func (p *servicePackage) Resources(_ context.Context) []servicepkg.ServicePackageResource {
	return []servicepkg.ServicePackageResource{
		{
			Factory: NewAppConfigResource,
		},
	}
}

func (p *servicePackage) DataSources(_ context.Context) []servicepkg.ServicePackageDataSource {
	return []servicepkg.ServicePackageDataSource{
		{
			Factory: NewAppConfigDataSource,
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return "app_config"
}
