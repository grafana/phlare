locals {
  name = "grafana-phlare-data-${random_id.suffix.hex}"
}

resource "random_id" "suffix" {
  byte_length = 4
}


resource "aws_s3_bucket" "bucket" {
  bucket = local.name
}

resource "aws_s3_bucket_acl" "bucket" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}
