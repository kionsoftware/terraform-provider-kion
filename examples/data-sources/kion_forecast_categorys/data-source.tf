data "kion_forecast_categorys" "all" {}

output "forecast_categories" {
  value = data.kion_forecast_categorys.all.data
}
