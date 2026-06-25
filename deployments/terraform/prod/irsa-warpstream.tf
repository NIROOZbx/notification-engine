# ==============================================================================
# WarpStream Agent IRSA Configurations
# ==============================================================================

# Trust relationship for the WarpStream Service Account (warpstream:warpstream-agent-sa)
data "aws_iam_policy_document" "warpstream_irsa_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_provider}:sub"
      values   = ["system:serviceaccount:warpstream:warpstream-agent-sa"]
    }

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_provider}:aud"
      values   = ["sts.amazonaws.com"]
    }

    principals {
      identifiers = [module.eks.oidc_provider_arn]
      type        = "Federated"
    }
  }
}

# The IAM Role for WarpStream
resource "aws_iam_role" "warpstream_role" {
  name               = "prod-warpstream-irsa-role"
  assume_role_policy = data.aws_iam_policy_document.warpstream_irsa_trust.json
}

# Define the S3 Access Policy
resource "aws_iam_policy" "warpstream_s3_policy" {
  name        = "prod-warpstream-s3-access-policy"
  description = "Grants WarpStream agents read/write access to the streaming S3 data bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::nirooz-warpstream-data-prod",
          "arn:aws:s3:::nirooz-warpstream-data-prod/*"
        ]
      }
    ]
  })
}

# Attach the S3 Access Policy to the WarpStream Role
resource "aws_iam_role_policy_attachment" "attach_warpstream_s3" {
  role       = aws_iam_role.warpstream_role.name
  policy_arn = aws_iam_policy.warpstream_s3_policy.arn
}

# Define the Secrets Manager Access Policy for WarpStream credentials
resource "aws_iam_policy" "warpstream_secrets_policy" {
  name        = "prod-warpstream-secrets-access-policy"
  description = "Grants WarpStream agents read access to their credentials in Secrets Manager"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          "arn:aws:secretsmanager:ap-south-1:711396988882:secret:warpstream/credentials-*"
        ]
      }
    ]
  })
}

# Attach the Secrets Manager Policy to the WarpStream Role
resource "aws_iam_role_policy_attachment" "attach_warpstream_secrets" {
  role       = aws_iam_role.warpstream_role.name
  policy_arn = aws_iam_policy.warpstream_secrets_policy.arn
}
