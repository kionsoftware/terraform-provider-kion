provider "kion" {
  # If these are commented out, they will be loaded from
  # environment variables.
  # url = "https://kion.example.com"
  # apikey = "key here"
}

# Create an IAM policy.
resource "kion_aws_iam_policy" "p1" {
  name         = "sample-resource"
  description  = "Provides AdministratorAccess to all AWS Services"
  aws_iam_path = ""
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "*",
            "Resource": "*"
        }
    ]
}
EOF
}

# Output the ID of the resource created.
output "policy_id" {
  value = kion_aws_iam_policy.p1.id
}
