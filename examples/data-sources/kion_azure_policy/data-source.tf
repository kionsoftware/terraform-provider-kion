# Example 1: List all Azure policies
data "kion_azure_policy" "all" {
}

# Example 2: Filter Azure policies by name
data "kion_azure_policy" "tag_policies" {
  filter {
    name   = "name"
    values = ["Require Environment Tag"]
  }
}

# Example 3: Filter Azure policies using regex
data "kion_azure_policy" "security_policies" {
  filter {
    name   = "name"
    values = [".*Security.*"]
    regex  = true
  }
}

# Example 4: Multiple filters
data "kion_azure_policy" "managed_storage" {
  filter {
    name   = "name"
    values = ["Storage"]
    regex  = true
  }
  filter {
    name   = "ct_managed"
    values = ["true"]
  }
}

# Example outputs
output "all_policies" {
  description = "List of all Azure policies"
  value       = data.kion_azure_policy.all.list
}

output "tag_policy_details" {
  description = "Details of tag policies"
  value       = data.kion_azure_policy.tag_policies.list
}

output "security_policy_count" {
  description = "Number of security-related policies"
  value       = length(data.kion_azure_policy.security_policies.list)
}

output "managed_storage_policies" {
  description = "List of managed storage policies"
  value       = data.kion_azure_policy.managed_storage.list
}