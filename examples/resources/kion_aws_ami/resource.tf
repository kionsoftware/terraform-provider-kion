resource "kion_aws_ami" "example" {
  account_id         = 123456789
  aws_ami_id         = "ami-0abcdef1234567890"
  name               = "My Example AMI"
  region             = "us-west-2"
  description        = "This is an example AMI"
  expires_at         = "2024-12-31T23:59:59Z" # Optional: set an expiration date
  sync_deprecation   = true                   # Optional: synchronize deprecation status
  sync_tags          = true                   # Optional: synchronize tags
  unavailable_in_aws = false                  # Optional: indicate if AMI is unavailable in AWS

  # Optional: Define owner user groups
  owner_user_groups = [
    {
      id = 1234
    }
  ]

  # Optional: Define owner users
  owner_users = [
    {
      id = 5678
    }
  ]
}

output "ami_id" {
  value = kion_aws_ami.example.id
}

output "ami_name" {
  value = kion_aws_ami.example.name
}
