resource "aws_ecr_repository" "backend" {
  name                 = "envoy-backend"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Environment = "production"
    Project     = "envoy"
  }
}

resource "aws_ecr_repository" "billing" {
  name                 = "billing-service"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Environment = "production"
    Project     = "envoy"
  }
}
