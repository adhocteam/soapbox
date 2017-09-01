output "blue_website_bucket_id" {
  value = "${aws_s3_bucket.blue_bucket.id}"
}

output "green_website_bucket_id" {
  value = "${aws_s3_bucket.green_bucket.id}"
}

output "blue_website_bucket_arn" {
  value = "${aws_s3_bucket.blue_bucket.arn}"
}

output "green_website_bucket_arn" {
  value = "${aws_s3_bucket.green_bucket.arn}"
}
