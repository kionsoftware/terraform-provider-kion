# Create a CloudFormation template.
resource "kion_aws_cloudformation_template" "t1" {
  name    = "sample-resource"
  regions = ["us-east-1"]
  # description = "Creates a test IAM role."
  # region                 = ""
  # sns_arns               = ""
  # template_parameters    = ""
  # termination_protection = false
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "Creates a test IAM role.",
    "Metadata": {
        "VersionDate": {
            "Value": "20180718"
        },
        "Identifier": {
            "Value": "blank-role.json"
        }
    },
    "Resources": {
        "EnvTestRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "RoleName": "env-test-role",
                "Path": "/",
                "Policies": [],
                "AssumeRolePolicyDocument": {
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "Service": [
                                    "ec2.amazonaws.com"
                                ]
                            },
                            "Action": [
                                "sts:AssumeRole"
                            ]
                        }
                    ]
                }
            }
        }
    }
}
EOF
  tags = {
    Owner      = "jdoe"
    Department = "DevOps"
  }
}

# Output the ID of the resource created.
output "template_id" {
  value = kion_aws_cloudformation_template.t1.id
}

# Example 1: Basic VPC CloudFormation template
resource "kion_aws_cloudformation_template" "basic_vpc" {
  name        = "basic-vpc-template"
  description = "Creates a basic VPC with public and private subnets"
  regions     = ["us-east-1"]

  owner_users { id = 1 }

  policy = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description = "Basic VPC with public and private subnets"
    Parameters = {
      VpcCidr = {
        Type = "String"
        Default = "10.0.0.0/16"
        Description = "CIDR block for the VPC"
      }
    }
    Resources = {
      VPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = { Ref = "VpcCidr" }
          EnableDnsHostnames = true
          EnableDnsSupport = true
          Tags = [
            {
              Key = "Name"
              Value = "Basic VPC"
            }
          ]
        }
      }
    }
  })
}

# Example 2: Complete template with all available options
resource "kion_aws_cloudformation_template" "complete" {
  # Required fields
  name    = "complete-template"
  regions = ["us-east-1", "us-west-2"]
  policy  = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description = "Complete template example"
    Parameters = {
      Environment = {
        Type = "String"
        Default = "dev"
        AllowedValues = ["dev", "staging", "prod"]
      }
    }
    Resources = {
      S3Bucket = {
        Type = "AWS::S3::Bucket"
        Properties = {
          BucketName = { "Fn::Join" : ["-", ["example-bucket", { Ref = "Environment" }]] }
        }
      }
    }
  })

  # Optional fields
  description = "Template demonstrating all available options"
  region      = "us-east-1"  # Default region

  # SNS notifications
  sns_arns = jsonencode([
    "arn:aws:sns:us-east-1:123456789012:stack-notifications"
  ])

  # Template parameters
  template_parameters = jsonencode({
    Environment = "dev"
  })

  # Enable termination protection
  termination_protection = true

  # Owner configuration
  owner_users { id = 1 }
  owner_user_groups { id = 2 }

  # Stack tags
  tags = {
    Environment = "Production"
    Department  = "DevOps"
    Owner       = "Team-A"
  }
}

# Example 3: S3 bucket with lifecycle rules
resource "kion_aws_cloudformation_template" "s3_lifecycle" {
  name        = "s3-lifecycle-template"
  description = "S3 bucket with lifecycle rules"
  regions     = ["us-east-1"]
  owner_users { id = 1 }

  policy = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description = "S3 bucket with lifecycle rules"
    Resources = {
      ArchiveBucket = {
        Type = "AWS::S3::Bucket"
        Properties = {
          LifecycleConfiguration = {
            Rules = [
              {
                Id = "ArchiveRule"
                Status = "Enabled"
                Transitions = [
                  {
                    StorageClass = "GLACIER"
                    TransitionInDays = 90
                  }
                ]
              }
            ]
          }
        }
      }
    }
  })
}

# Example 4: Template with multiple owner groups
resource "kion_aws_cloudformation_template" "multi_owner" {
  name        = "multi-owner-template"
  description = "Template with multiple owner groups"
  regions     = ["us-east-1"]

  owner_user_groups {
    id = 1
  }
  owner_user_groups {
    id = 2
  }
  owner_user_groups {
    id = 3
  }

  policy = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Description = "Simple EC2 instance"
    Resources = {
      Instance = {
        Type = "AWS::EC2::Instance"
        Properties = {
          InstanceType = "t2.micro"
          ImageId = "ami-0123456789abcdef0"
        }
      }
    }
  })
}

# Output examples
output "basic_template_id" {
  description = "ID of the basic VPC template"
  value = kion_aws_cloudformation_template.basic_vpc.id
}

output "complete_template_id" {
  description = "ID of the complete template"
  value = kion_aws_cloudformation_template.complete.id
}

output "template_regions" {
  description = "Regions where the complete template is available"
  value = kion_aws_cloudformation_template.complete.regions
}
