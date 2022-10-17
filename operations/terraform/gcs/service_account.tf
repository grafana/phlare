resource "google_service_account" "user" {
  account_id   = local.name
  display_name = "My Service Account"
}

resource "google_service_account_key" "user" {
  service_account_id = google_service_account.user.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}

resource "google_storage_bucket_iam_member" "user" {
  bucket = google_storage_bucket.bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.user.email}"
}
