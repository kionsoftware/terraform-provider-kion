data "kion_label" "cost_center_x" {
  filter {
    name   = "key"
    values = ["cost_center"]
  }
  filter {
    name   = "value"
    values = ["x"]
  }
}
