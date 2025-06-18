# Create an OCI Commercial billing source
resource "kion_billing_source_oci" "example" {
  name                    = "My OCI Billing Source"
  billing_start_date      = "2024-01"
  account_type_id         = 26  # 26 = OCI Commercial, 27 = OCI Government, 28 = OCI Federal
  
  # OCI API access credentials
  tenancy_ocid            = "ocid1.tenancy.oc1..exampleuniqueID"
  user_ocid               = "ocid1.user.oc1..exampleuniqueID"
  fingerprint             = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
  private_key             = file("~/.oci/oci_api_key.pem")
  region                  = "us-ashburn-1"
  
  # Billing configuration
  is_parent_tenancy       = true
  use_focus_reports       = true
  use_proprietary_reports = true
}

# Create an OCI Government billing source
resource "kion_billing_source_oci" "gov_example" {
  name                    = "My OCI Gov Billing Source"
  billing_start_date      = "2024-01"
  account_type_id         = 27  # OCI Government
  
  tenancy_ocid            = "ocid1.tenancy.oc1..govexampleuniqueID"
  user_ocid               = "ocid1.user.oc1..govexampleuniqueID"
  fingerprint             = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
  private_key             = var.oci_private_key  # Can use a variable for sensitive data
  region                  = "us-luke-1"  # Government region
  
  is_parent_tenancy       = false
  use_focus_reports       = true
  use_proprietary_reports = false
  
  # Skip validation during creation if needed
  skip_validation         = false
}