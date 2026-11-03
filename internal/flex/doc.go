// Package flex provides helper functions for converting between Terraform
// Plugin Framework types (types.String, types.Int64, etc.) and the Kion SDK
// generated types (string, int64, generated.OptString, etc.).
//
// "Expand" functions convert from Terraform → SDK (used in Create/Update).
// "Flatten" functions convert from SDK → Terraform (used in Read).
package flex
