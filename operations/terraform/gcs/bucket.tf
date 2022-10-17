locals {
  name = "grafana-phlare-data-${random_id.suffix.hex}"
}

resource "random_id" "suffix" {
  byte_length = 4
}

resource "google_storage_bucket" "bucket" {
  name          = local.name
  location      = "EU"
  force_destroy = true

  uniform_bucket_level_access = true
}

