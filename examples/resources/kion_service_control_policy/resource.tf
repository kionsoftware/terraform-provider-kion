# Create a deny policy for critical service modifications
resource "kion_service_control_policy" "protect_critical_services" {
  name        = "Protect Critical Services"
  description = "Prevents modification of critical AWS services and configurations"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "config:DeleteConfigurationRecorder",
        "config:StopConfigurationRecorder",
        "cloudtrail:StopLogging",
        "cloudtrail:DeleteTrail",
        "guardduty:DeleteDetector",
        "guardduty:DisassociateFromMasterAccount"
      ],
      "Resource": "*"
    }
  ]
}
EOF

  # Policy owners
  owner_users {
    id = 10  # Security Lead
  }
  owner_user_groups {
    id = 20  # Security Team
  }
}

# Create a policy to enforce tagging standards
resource "kion_service_control_policy" "enforce_tagging" {
  name        = "Mandatory Resource Tagging"
  description = "Enforces mandatory tags on all resources"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "ec2:RunInstances",
        "rds:CreateDBInstance",
        "s3:CreateBucket"
      ],
      "Resource": "*",
      "Condition": {
        "Null": {
          "aws:RequestTag/Environment": "true",
          "aws:RequestTag/CostCenter": "true",
          "aws:RequestTag/Owner": "true"
        }
      }
    }
  ]
}
EOF

  owner_users {
    id = 11  # Platform Lead
  }
  owner_user_groups {
    id = 21  # Platform Team
  }
}

# Create a policy to restrict regions
resource "kion_service_control_policy" "region_restriction" {
  name        = "Region Restriction"
  description = "Restricts AWS resource creation to specific regions"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "NotAction": [
        "cloudfront:*",
        "iam:*",
        "route53:*",
        "support:*"
      ],
      "Resource": "*",
      "Condition": {
        "StringNotEquals": {
          "aws:RequestedRegion": [
            "us-east-1",
            "us-east-2",
            "us-west-2"
          ]
        }
      }
    }
  ]
}
EOF

  owner_users {
    id = 12  # Cloud Security Architect
  }
  owner_user_groups {
    id = 22  # Cloud Architecture Team
  }
}

# Output policy information
output "service_control_policies" {
  value = {
    protect_critical = {
      id                   = kion_service_control_policy.protect_critical_services.id
      created_by_user_id   = kion_service_control_policy.protect_critical_services.created_by_user_id
      system_managed      = kion_service_control_policy.protect_critical_services.system_managed_policy
    }
    enforce_tagging = {
      id                   = kion_service_control_policy.enforce_tagging.id
      created_by_user_id   = kion_service_control_policy.enforce_tagging.created_by_user_id
      system_managed      = kion_service_control_policy.enforce_tagging.system_managed_policy
    }
    region_restriction = {
      id                   = kion_service_control_policy.region_restriction.id
      created_by_user_id   = kion_service_control_policy.region_restriction.created_by_user_id
      system_managed      = kion_service_control_policy.region_restriction.system_managed_policy
    }
  }
  description = "Details of created service control policies"
}
