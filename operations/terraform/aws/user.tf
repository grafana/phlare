resource "aws_iam_access_key" "bucket" {
  user = aws_iam_user.bucket.name
}

resource "aws_iam_user" "bucket" {
  name = local.name
}

resource "aws_iam_user_policy" "bucket_rw" {
  name = local.name
  user = aws_iam_user.bucket.name

  policy = data.aws_iam_policy_document.bucket_rw.json
}

data "aws_iam_policy_document" "bucket_rw" {
  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.bucket.arn,
    ]
  }
  statement {
    actions = [
      "s3:*Object",
    ]

    resources = [
      "${aws_s3_bucket.bucket.arn}/*",
    ]
  }
}
