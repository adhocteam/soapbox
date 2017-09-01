provider "aws" {
  alias  = "${var.region}"
  region = "${var.region}"
}

data "template_file" "blue_bucket_policy" {
  template = "${file("${path.module}/website_bucket_policy.json")}"

  vars {
    bucket = "${var.application_name}-${var.environment}-blue-deploy-bucket"
  }
}

data "template_file" "green_bucket_policy" {
  template = "${file("${path.module}/website_bucket_policy.json")}"

  vars {
    bucket = "${var.application_name}-${var.environment}-green-deploy-bucket"
  }
}

resource "aws_s3_bucket" "blue_bucket" {
  provider = "aws.${var.region}"
  bucket   = "${var.application_name}-${var.environment}-blue-deploy-bucket"
  policy   = "${data.template_file.blue_bucket_policy.rendered}"

  website {
    index_document = "index.html"
    error_document = "404.html"
  }

  tags = "${merge("${var.tags}",map("Name", "${var.application_name}-${var.environment}.${var.platform_domain}", "env", "${var.environment}", "app", "${var.application_name}"))}"

}

resource "aws_s3_bucket" "green_bucket" {
  provider = "aws.${var.region}"
  bucket   = "${var.application_name}-${var.environment}-green-deploy-bucket"
  policy   = "${data.template_file.green_bucket_policy.rendered}"

  website {
    index_document = "index.html"
    error_document = "404.html"
  }

  tags = "${merge("${var.tags}",map("Name", "${var.application_name}-${var.environment}.${var.platform_domain}", "env", "${var.environment}", "app", "${var.application_name}"))}"

}

resource "aws_cloudfront_distribution" "website_cdn" {
  enabled      = true
  price_class  = "PriceClass_200"
  http_version = "http1.1"

  "origin" {
    origin_id   = "origin-bucket-${aws_s3_bucket.blue_bucket.id}"
    domain_name = "${aws_s3_bucket.blue_bucket.website_endpoint}"

    custom_origin_config {
      origin_protocol_policy = "http-only"
      http_port              = "80"
      https_port             = "443"
      origin_ssl_protocols   = ["TLSv1"]
    }

  }

  default_root_object = "index.html"

  custom_error_response {
    error_code            = "404"
    error_caching_min_ttl = "360"
    response_code         = "200"
    response_page_path    = "${var.not-found-response-path}"
  }

  "default_cache_behavior" {
    allowed_methods = ["GET", "HEAD", "DELETE", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods  = ["GET", "HEAD"]

    "forwarded_values" {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    trusted_signers = ["${var.trusted_signers}"]

    min_ttl          = "0"
    default_ttl      = "300"                                              //3600
    max_ttl          = "1200"                                             //86400
    target_origin_id = "origin-bucket-${aws_s3_bucket.blue_bucket.id}"

    // This redirects any HTTP request to HTTPS. Security first!
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
  }

  "restrictions" {
    "geo_restriction" {
      restriction_type = "none"
    }
  }

  "viewer_certificate" {
    cloudfront_default_certificate = true
  }

  aliases = ["${var.application_name}-${var.environment}.${var.platform_domain}"]

  tags = "${merge("${var.tags}",map("Name", "${var.application_name}-${var.environment}.${var.platform_domain}", "env", "${var.environment}", "app", "${var.application_name}"))}"

}

data "aws_route53_zone" "platform_zone" {
  name = "${var.platform_domain}."
}

resource "aws_route53_record" "application_subdomain" {
  zone_id = "${data.aws_route53_zone.platform_zone.zone_id}"
  name    = "${var.application_name}.${var.environment}"
  type    = "CNAME"
  ttl     = "300"

  records = ["${aws_cloudfront_distribution.website_cdn.domain_name}"]
}
