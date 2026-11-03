package custom_variable_override

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
			Factory: NewCustomVariableOverrideResource,
		},
	}
}

func (p *servicePackage) DataSources(_ context.Context) []servicepkg.ServicePackageDataSource {
	return []servicepkg.ServicePackageDataSource{
		{
			Factory: NewCustomVariableOverrideDataSource,
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return "custom_variable_override"
}
