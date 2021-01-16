locals {
  folder_id = 507359599160

  services = [
    "appengine.googleapis.com",
    "run.googleapis.com",
    "iap.googleapis.com",
  ]
  location = "europe-west1"

  domain = terraform.workspace == "prod" ? "weight.${var.domain}" : "${terraform.workspace}.weight.${var.domain}"
}
