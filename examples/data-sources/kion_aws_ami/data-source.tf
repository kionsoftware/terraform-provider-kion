data "kion_aws_ami" "example" {
  filter {
    name   = "region"
    values = ["us-west-2"]
  }

  filter {
    name   = "name"
    values = ["^MyExampleAMI.*"] # Use regex if you want to match a pattern
    regex  = true
  }
}

output "ami_list" {
  value = data.kion_aws_ami.example.list
}

output "first_ami_id" {
  value = data.kion_aws_ami.example.list[0].aws_ami_id
}

output "first_ami_name" {
  value = data.kion_aws_ami.example.list[0].name
}
