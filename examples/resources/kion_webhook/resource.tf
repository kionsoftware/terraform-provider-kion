resource "kion_webhook" "example" {
  # Required
  callout_url        = "example"
  name               = "example"
  timeout_in_seconds = 1

  # Optional
  # description             = "example"
  # owner_user_group_ids    = []
  # owner_user_ids          = []
  # request_body            = "example"
  # request_headers         = "example"
  # request_method          = "example"
  # should_send_secure_info = false
  # skip_ssl                = false
  # use_request_headers     = false
}
