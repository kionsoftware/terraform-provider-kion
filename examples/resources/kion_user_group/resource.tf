resource "kion_user_group" "example" {
  # Required
  idms_id = 1
  name    = "example"

  # Optional
  # add_self_as_viewer    = false
  # description           = "example"
  # last_updated          = "example"
  # owner_user_group_ids  = []
  # owner_user_ids        = []
  # user_ids              = []
  # viewer_user_group_ids = []
  # viewer_user_ids       = []
}
