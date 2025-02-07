# Find an AWS managed policy to reference
data "kion_aws_iam_policy" "aws_policy" {
  query       = "AdministratorAccess"
  policy_type = "aws"
}

# Create an IAM policy
resource "kion_aws_iam_policy" "p1" {
  name         = "sample-resource"
  description  = "Based on AWS managed policy"
  aws_iam_path = ""
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = data.kion_aws_iam_policy.aws_policy.list[0].policy
}

# Output the ID of the resource created
output "policy_id" {
  value = kion_aws_iam_policy.p1.id
}
