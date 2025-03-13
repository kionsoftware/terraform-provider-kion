# Example 1: Create a basic IAM policy
resource "kion_aws_iam_policy" "basic" {
  name        = "basic-iam-policy"
  description = "Basic IAM policy example"
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObject"
        ]
        Resource = [
          "arn:aws:s3:::example-bucket",
          "arn:aws:s3:::example-bucket/*"
        ]
      }
    ]
  })
  owner_users { id = 1 }
}

# Example 2: Create an IAM policy with all available options
resource "kion_aws_iam_policy" "complete" {
  # Required fields
  name        = "complete-iam-policy"
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "*"
        Resource = "*"
      }
    ]
  })

  # Optional fields
  description  = "Complete IAM policy example with all options"
  aws_iam_path = "/custom/path/"

  # Owner configuration - must provide at least one of owner_users or owner_user_groups
  owner_users {
    id = 1
  }
  owner_user_groups {
    id = 2
  }
}

# Example 3: Create an IAM policy based on an AWS managed policy
data "kion_aws_iam_policy" "aws_policy" {
  query       = "ReadOnlyAccess"
  policy_type = "aws"
}

resource "kion_aws_iam_policy" "from_aws" {
  name        = "aws-readonly-based"
  description = "Based on AWS ReadOnlyAccess managed policy"
  policy      = data.kion_aws_iam_policy.aws_policy.list[0].policy
  owner_users { id = 1 }
}

# Example 4: Create an IAM policy with multiple owner groups
resource "kion_aws_iam_policy" "multi_owner" {
  name        = "multi-owner-policy"
  description = "Policy with multiple owner groups"
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:Describe*"
        ]
        Resource = "*"
      }
    ]
  })

  owner_user_groups {
    id = 1
  }
  owner_user_groups {
    id = 2
  }
  owner_user_groups {
    id = 3
  }
}

# Output examples
output "basic_policy_id" {
  value = kion_aws_iam_policy.basic.id
}

output "complete_policy_id" {
  value = kion_aws_iam_policy.complete.id
}

output "aws_based_policy_id" {
  value = kion_aws_iam_policy.from_aws.id
}
