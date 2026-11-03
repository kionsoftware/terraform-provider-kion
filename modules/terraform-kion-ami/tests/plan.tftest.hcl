# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    account_id     = 1
    aws_ami_id     = "example"
    name           = "example"
    owner_user_ids = []
    region         = "example"
  }
}
