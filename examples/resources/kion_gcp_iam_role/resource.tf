# Create a complete GCP IAM role with all available options
resource "kion_gcp_iam_role" "complete_example" {
  name                  = "Custom Admin Role"
  description          = "Comprehensive administrative role for GCP resources"
  gcp_role_launch_stage = 4  # GA (Generally Available)
  system_managed_policy = false

  # Role permissions
  role_permissions = [
    "compute.instances.list",
    "compute.instances.get",
    "compute.instances.start",
    "compute.instances.stop",
    "compute.disks.list",
    "compute.disks.get",
    "storage.buckets.list",
    "storage.objects.list"
  ]

  # Ownership configuration
  owner_users {
    id = 1  # Platform Admin
  }
  owner_users {
    id = 2  # Security Admin
  }

  owner_user_groups {
    id = 1  # Cloud Platform Team
  }
}

# Create a read-only role
resource "kion_gcp_iam_role" "readonly_example" {
  name                  = "Read Only Access"
  description          = "Provides read-only access to GCP resources"
  gcp_role_launch_stage = 4

  role_permissions = [
    "compute.instances.list",
    "compute.instances.get",
    "compute.disks.list",
    "compute.disks.get",
    "storage.buckets.list",
    "storage.objects.list"
  ]

  owner_users {
    id = 1
  }
}

# Create a storage admin role
resource "kion_gcp_iam_role" "storage_admin" {
  name                  = "Storage Administrator"
  description          = "Manages GCP storage resources"
  gcp_role_launch_stage = 4

  role_permissions = [
    "storage.buckets.create",
    "storage.buckets.delete",
    "storage.buckets.get",
    "storage.buckets.list",
    "storage.buckets.update",
    "storage.objects.create",
    "storage.objects.delete",
    "storage.objects.get",
    "storage.objects.list",
    "storage.objects.update"
  ]

  owner_user_groups {
    id = 2  # Storage Team
  }
}

# Output examples
output "admin_role_id" {
  value = kion_gcp_iam_role.complete_example.id
}

output "admin_role_gcp_id" {
  value = kion_gcp_iam_role.complete_example.gcp_id
}

output "readonly_role_id" {
  value = kion_gcp_iam_role.readonly_example.id
}

output "storage_role_id" {
  value = kion_gcp_iam_role.storage_admin.id
}
