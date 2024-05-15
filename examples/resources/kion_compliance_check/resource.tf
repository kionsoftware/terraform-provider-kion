resource "kion_compliance_check" "c1" {
  name                     = "sample-resource"
  cloud_provider_id        = 1
  compliance_check_type_id = 1
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  #   body = <<EOF
  # {
  #     "Version": "2012-10-17",
  #     "Statement": [
  #         {
  #             "Effect": "Allow",
  #             "Action": "*",
  #             "Resource": "*"
  #         }
  #     ]
  # }
  # EOF
}

# Output the ID of the resource created.
output "check_id" {
  value = kion_compliance_check.c1.id
}
