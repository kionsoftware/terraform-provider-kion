resource "kion_iam_policy" "example" {
  # Required
  name   = "example"
  policy = "example"

  # Optional
  # aws_iam_path                  = "example"
  # car_restricted                = false
  # car_restricted_user_group_ids = []
  # car_restricted_user_ids       = []
  # description                   = "example"
  # last_updated                  = "example"
  # owner_user_group_ids          = []
  # owner_user_ids                = []
}
