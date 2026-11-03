resource "kion_oauth2_client" "example" {
  # Required
  grant_types = []
  name        = "example"

  # Optional
  # allowed_scopes          = []
  # description             = "example"
  # is_public               = false
  # owner_user_id           = 1
  # redirect_uris           = []
  # service_account_user_id = 1
}
