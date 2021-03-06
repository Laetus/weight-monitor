resource "random_pet" "project_id" {
  length = 3
}

resource "google_project" "this" {
  name            = "weight-monitor-${terraform.workspace}"
  project_id      = substr(random_pet.project_id.id, 0, 20)
  folder_id       = local.folder_id
  billing_account = var.billing_account
}

resource "google_project_service" "this" {
  for_each = toset(local.services)
  project  = google_project.this.project_id
  service  = each.value

  disable_dependent_services = true
}

resource "google_project_iam_member" "cloud_build" {
  project = google_project.this.project_id
  role    = "roles/editor"
  member  = "serviceAccount:${var.cloud_build_service_account}"
}

resource "google_iap_brand" "this" {
  support_email     = var.support_email
  application_title = "Laetus Inc."
  project           = google_project.this.project_id
}

// Setup firestore instance
resource "google_app_engine_application" "this" {
  project       = google_project.this.project_id
  location_id   = "europe-west"
  database_type = "CLOUD_FIRESTORE"
}


