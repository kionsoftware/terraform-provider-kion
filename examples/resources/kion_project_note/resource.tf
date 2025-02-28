# Create project documentation note
resource "kion_project_note" "project_documentation" {
  name           = "Project Overview and Configuration"
  project_id     = 10
  create_user_id = 15  # DevOps Lead
  text           = <<-EOT
    # Project Overview
    This project contains infrastructure for the development environment.

    ## Important Contacts
    - Project Owner: Jane Smith (jane.smith@company.com)
    - Technical Lead: John Doe (john.doe@company.com)
    - Security Contact: Sarah Johnson (sarah.johnson@company.com)

    ## Key Information
    - Environment: Development
    - Cost Center: IT-1234
    - Team: Platform Engineering
    - Start Date: 2024-01-01

    ## Infrastructure Details
    ### AWS Resources
    - Region: us-east-1
    - VPC ID: vpc-abc123def456
    - Subnet IDs: subnet-123, subnet-456
    - Security Groups: sg-789, sg-012

    ### Azure Resources
    - Region: eastus
    - Resource Group: dev-platform-rg
    - VNET: platform-vnet

    ## Security Requirements
    1. All resources must be tagged
    2. No public-facing resources without security review
    3. Daily backups required for critical systems
    4. All changes must go through IaC
  EOT
}

# Create operational runbook note
resource "kion_project_note" "operational_runbook" {
  name           = "Operational Runbook"
  project_id     = 10
  create_user_id = 16  # Operations Engineer
  text           = <<-EOT
    # Operational Runbook

    ## Daily Checks
    1. Review CloudWatch dashboards
    2. Check backup completion status
    3. Verify auto-scaling metrics
    4. Monitor cost alerts

    ## Common Issues and Resolution
    ### High CPU Usage
    1. Check CloudWatch metrics
    2. Review application logs
    3. Scale up if necessary

    ### Network Connectivity
    1. Verify security group rules
    2. Check VPC flow logs
    3. Validate route tables

    ## Emergency Procedures
    ### Production Access
    1. Request elevation through Kion
    2. Use break-glass procedure
    3. Document all actions taken

    ### System Recovery
    1. Identify failure point
    2. Execute relevant playbook
    3. Update status page
  EOT
}

# Create compliance documentation note
resource "kion_project_note" "compliance_doc" {
  name           = "Compliance Requirements"
  project_id     = 10
  create_user_id = 17  # Compliance Officer
  text           = <<-EOT
    # Compliance Documentation

    ## Regulatory Requirements
    - SOC 2 Type II
    - HIPAA
    - PCI DSS

    ## Data Classification
    - PII: Restricted
    - Financial: Confidential
    - Public: Unrestricted

    ## Access Control
    1. Role-based access control (RBAC)
    2. Multi-factor authentication (MFA)
    3. Regular access reviews

    ## Audit Requirements
    - Monthly access reviews
    - Quarterly compliance scans
    - Annual penetration testing

    ## Incident Response
    1. Security incident reporting
    2. Data breach notification
    3. Recovery procedures
  EOT
}

# Output note information
output "project_notes" {
  value = {
    documentation = {
      id         = kion_project_note.project_documentation.id
      created_at = kion_project_note.project_documentation.created_at
      creator    = kion_project_note.project_documentation.create_user_name
    }
    runbook = {
      id         = kion_project_note.operational_runbook.id
      created_at = kion_project_note.operational_runbook.created_at
      creator    = kion_project_note.operational_runbook.create_user_name
    }
    compliance = {
      id         = kion_project_note.compliance_doc.id
      created_at = kion_project_note.compliance_doc.created_at
      creator    = kion_project_note.compliance_doc.create_user_name
    }
  }
  description = "Details of created project notes"
}
