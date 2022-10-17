output "extra_flags" {
  sensitive = true
  value = [
    "-storage.backend=gcs",
    "-storage.gcs.bucket-name=${local.name}",
    "-storage.gcs.service-account=${jsonencode(jsondecode(base64decode(google_service_account_key.user.private_key)))}",
  ]
}

output "service_account" {
  sensitive = true
  value     = base64decode(google_service_account_key.user.private_key)
}
