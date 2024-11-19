resource "kion_aws_ami" "example" {
  account_id                = 1                          # Required: AWS account application ID where the AMI is stored
  aws_ami_id                = "ami-123456"               # Required: Image ID of the AMI from AWS
  description               = "Gold image for RHEL 7.5." # Optional: Description for the AMI in the application
  expiration_alert_number   = 1                          # Optional: The amount of time before the expiration alert is shown
  expiration_alert_unit     = "days"                     # Optional: The unit of time for the expiration alert (e.g., 'days', 'hours')
  expiration_notify         = true                       # Optional: Will notify the owners that the shared AMI has expired
  expiration_warning_number = 1                          # Optional: The amount of time before the expiration warning is sent
  expiration_warning_unit   = "days"                     # Optional: The unit of time for the expiration warning (e.g., 'days', 'hours')
  expires_at                = "2024-12-31T22:10:41.406Z" # Optional: Set an expiration date
  name                      = "rhel-7-5-20180213"        # Required: The name of the AMI
  owner_user_group_ids      = [1, 2]                     # Optional: List of group IDs who will own the AMI
  owner_user_ids            = [1, 2]                     # Optional: List of user IDs who will own the AMI
  region                    = "us-east-1"                # Required: AWS region where the AMI exists
  sync_deprecation          = true                       # Optional: Will sync the expiration date from the system into the AMI in AWS
  sync_tags                 = true                       # Optional: Will sync the AWS tags from the source AMI into all the accounts where the AMI is shared
}

output "ami_id" {
  value = kion_aws_ami.example.id
}

output "ami_name" {
  value = kion_aws_ami.example.name
}
