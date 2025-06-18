# Create an OU Cloud Access Role Exemption
resource "kion_ou_cloud_access_role_exemption" "example" {
  ou_cloud_access_role_id = 107
  ou_id                   = 104
  reason                  = "This CAR isn't used in this OU"
}