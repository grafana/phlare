locals {
  endpoint_split = split(".", aws_s3_bucket.bucket.bucket_regional_domain_name)
  endpoint       = join(".", slice(local.endpoint_split, 1, length(local.endpoint_split)))
}


data "aws_region" "current" {}

output "extra_flags" {
  sensitive = true
  value = [
    "-storage.backend=s3",
    "-storage.s3.bucket-name=${local.name}",
    "-storage.s3.access-key-id=${aws_iam_access_key.bucket.id}",
    "-storage.s3.secret-access-key=${aws_iam_access_key.bucket.secret}",
    "-storage.s3.region=${data.aws_region.current.name}",
    "-storage.s3.endpoint=${local.endpoint}",
  ]
}

