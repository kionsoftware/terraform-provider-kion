resource "kion_cft" "example" {
  # Required
  name    = "example"
  policy  = "example"
  regions = []

  # Optional
  # description            = "example"
  # owner_user_group_ids   = []
  # owner_user_ids         = []
  # region                 = "example"
  # sns_arns               = "example"
  # tags = [{
  #   tag_key   = "example"
  #   tag_value = "example"
  # }]
  # template_parameters    = "example"
  # termination_protection = false
}
