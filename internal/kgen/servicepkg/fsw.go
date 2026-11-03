package servicepkg

import "terraform-provider-kion/internal/kgen/kfs"

// fsw is the filesystem seam the generator writes through. Production uses the
// real OS; tests swap it for a mock (see the package's _test.go files).
var fsw kfs.FS = kfs.OS{}
