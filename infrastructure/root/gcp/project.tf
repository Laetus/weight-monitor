locals {
  folder_id = 507359599160

  services = [
    "run.googleapis.com"
  ]
}

resource "random_pet" "project_id" {
  length = 3
}

resource "google_project" "this" {
  name       = "weight-monitor-${terraform.workspace}"
  project_id = substr(random_pet.project_id.id, 0, 20)
  folder_id  = local.folder_id
  billing_account = var.billing_account
}

resource "google_project_service" "this" {
  for_each = toset(local.services)
  project  = google_project.this.project_id
  service  = each.value

  disable_dependent_services = true
}

