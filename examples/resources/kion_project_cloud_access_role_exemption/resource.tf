# Create a Project Cloud Access Role Exemption
resource "kion_project_cloud_access_role_exemption" "example" {
  ou_cloud_access_role_id = 107
  project_id              = 42
  reason                  = "This CAR isn't used in this project"
}