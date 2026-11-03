resource "kion_compliance_family" "example" {
  # Required
  compliance_program_id = 1
  name                  = "example"

  # Optional
  # description = "example"
}
