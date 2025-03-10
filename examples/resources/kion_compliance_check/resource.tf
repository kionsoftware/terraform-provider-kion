# AWS Compliance Check Example - IAM Certificate Expiration
resource "kion_compliance_check" "aws_iam_certificate" {
  name                     = "IAM Certificate Expiration Check"
  description              = "Identify expired IAM SSL/TLS certificates"
  cloud_provider_id        = 1 # AWS
  compliance_check_type_id = 1
  frequency_minutes        = 60
  frequency_type_id        = 3
  is_all_regions           = true
  is_auto_archived         = false
  severity_type_id         = 3
  created_by_user_id       = 1

  body = <<-EOT
---
policies:
  - name: iam-certificate-expired
    resource: aws.iam-certificate
    description: |
      Identify expired IAM SSL/TLS certificates
    filters:
      - type: value
        key: Expiration
        value_type: expiration
        op: lt
        value: 0
    actions:
      - type: webhook
        url: '{{CT::CallbackURL}}'
        method: POST
        batch: true
        headers:
          Authorization: '`{{CT::Authorization}}`'
        body: |-
          {
            "compliance_check_id": `{{CT::CheckId}}`,
            "account_number": account_id,
            "region": region,
            "scan_started_at": execution_start,
            "findings": resources[].{resource_name: ServerCertificateId, resource_type: `iam-certificate`, data_json: {key_create_date: "c7n:matched-keys"[].CreateDate}}
          }
EOT

  owner_users {
    id = 1
  }
}

# Azure Compliance Check Example - Network Security Group Port 3389
resource "kion_compliance_check" "azure_nsg_3389" {
  name                     = "NSG Port 3389 Check"
  description              = "Identify network security groups containing rules that allow ingress from port 3389"
  cloud_provider_id        = 2 # Azure
  compliance_check_type_id = 2
  azure_policy_id          = 1
  frequency_minutes        = 60
  frequency_type_id        = 3
  is_all_regions           = true
  is_auto_archived         = false
  severity_type_id         = 3

  body = <<-EOT
---
policies:
  - name: network-sg-with-inbound-3389
    resource: azure.networksecuritygroup
    description: |
      Identify network security groups containing rules that allow ingress from port 3389
    filters:
      - type: ingress
        ipProtocol: '*'
        ports: '3389'
        match: 'any'
        access: 'Allow'
    actions:
      - type: webhook
        url: '{{CT::CallbackURL}}'
        method: POST
        batch: true
        headers:
          Authorization: '`{{CT::Authorization}}`'
        body: |-
          {
            "compliance_check_id": `{{CT::CheckId}}`,
            "account_number": account_id,
            "scan_started_at": execution_start,
            "findings": resources[].{region: location, resource_name: name, resource_type: `Microsoft.Network/networkSecurityGroups`}
          }
EOT

  owner_user_groups {
    id = 1
  }
}

# GCP Compliance Check Example - Compute Instance Tags
resource "kion_compliance_check" "gcp_compute_tags" {
  name                     = "Compute Instance Tag Compliance"
  description              = "Identify Compute instances that are not compliant with tagging policies"
  cloud_provider_id        = 3 # GCP
  compliance_check_type_id = 3
  frequency_minutes        = 60
  frequency_type_id        = 3
  is_all_regions           = true
  is_auto_archived         = false
  severity_type_id         = 3

  body = <<-EOT
---
vars:
  absent-tags-filter:
    or:
      - "tag:data_class": absent
      - "tag:owner": absent
      - "tag:service": absent

policies:
  - name: compute-instance-with-non-compliant-tags
    resource: gcp.instance
    description: |
      Identify Compute instances that are not compliant with tagging policies
    filters:
      - absent-tags-filter
    actions:
      - type: webhook
        url: '{{CT::CallbackURL}}'
        method: POST
        batch: true
        headers:
          Authorization: '`{{CT::Authorization}}`'
        body: |-
          {
            "compliance_check_id": `{{CT::CheckId}}`,
            "account_number": account_id,
            "scan_started_at": execution_start,
            "findings": resources[].{resource_name: name, resource_type: kind, region: zone},
            "data_json": {
              "instance_id": id
            }
          }
EOT

  owner_users {
    id = 1
  }
}

# Output examples
output "aws_check_id" {
  value = kion_compliance_check.aws_iam_certificate.id
}

output "azure_check_id" {
  value = kion_compliance_check.azure_nsg_3389.id
}

output "gcp_check_id" {
  value = kion_compliance_check.gcp_compute_tags.id
}

output "last_scan_id" {
  value = kion_compliance_check.aws_iam_certificate.last_scan_id
}
