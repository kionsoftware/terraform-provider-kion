# Create an AWS Service Control Policy.
resource "kion_service_control_policy" "scp1" {
  name        = "Test SCP"
  description = "This is a sample SCP."
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "config:StopConfigurationRecorder"
      ],
      "Resource": "*"
    }
  ]
}
EOF
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "scp_id" {
  value = kion_service_control_policy.scp1.id
}
