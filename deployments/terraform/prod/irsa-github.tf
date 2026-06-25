# ==============================================================================
# GitHub Actions CI/CD OIDC Roles
# ==============================================================================

# Create the OIDC provider for GitHub Actions
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1", "1c58a3a8518e8759bf075b76b750d4f2df264fcd"]
}

# ----------------------------------------------------
# 1. envoy-backend Deployer Role
# ----------------------------------------------------

# Trust policy permitting GitHub Actions OIDC to assume the role
data "aws_iam_policy_document" "github_oidc_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      # Limits role assumption strictly to your notification-engine repository
      values   = ["repo:NIROOZbx/notification-engine:*"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.github.arn]
      type        = "Federated"
    }
  }
}

# The IAM Role for GitHub Actions Deployer (envoy-backend)
resource "aws_iam_role" "github_actions_role" {
  name               = "github-actions-deploy-role"
  assume_role_policy = data.aws_iam_policy_document.github_oidc_trust.json
}

# Attach ECR push permissions to the GitHub Actions Role
resource "aws_iam_role_policy_attachment" "attach_github_ecr" {
  role       = aws_iam_role.github_actions_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
}

# ----------------------------------------------------
# 2. billing-service Deployer Role
# ----------------------------------------------------

# Trust policy permitting GitHub Actions OIDC for billing-service to assume the role
data "aws_iam_policy_document" "billing_oidc_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      # Limits role assumption strictly to your billing-service repository
      values   = ["repo:NIROOZbx/billing-service:*"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.github.arn]
      type        = "Federated"
    }
  }
}

# The IAM Role for GitHub Actions Deployer (billing-service)
resource "aws_iam_role" "billing_github_actions_role" {
  name               = "github-actions-billing-deploy-role"
  assume_role_policy = data.aws_iam_policy_document.billing_oidc_trust.json
}

# Attach ECR push permissions to the Billing Actions Role
resource "aws_iam_role_policy_attachment" "attach_billing_github_ecr" {
  role       = aws_iam_role.billing_github_actions_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
}
