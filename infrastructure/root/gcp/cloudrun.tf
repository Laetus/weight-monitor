resource "google_cloud_run_service" "this" {
  name     = "weight-monitor"
  location = local.location
  project  = google_project.this.project_id

  metadata {
    namespace = google_project.this.project_id
  }

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello"
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template,
      metadata,
    ]

  }
}

resource "google_cloud_run_domain_mapping" "this" {
  location = local.location
  name     = local.domain
  project  = google_project.this.project_id

  metadata {
    namespace = google_project.this.project_id
  }

  spec {
    route_name = google_cloud_run_service.this.name
  }
}

// This service account needs read permissions on the Container Registry Bucket otherwise the cloud
// run deployments fails
output "cloud_run_service_account" {
  value = "service-${google_project.this.number}@serverless-robot-prod.iam.gserviceaccount.com"
}

output "cloud_run_status" {
  value = element(google_cloud_run_service.this.status, 0)["url"]
}
