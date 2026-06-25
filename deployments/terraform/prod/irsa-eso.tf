# ==============================================================================
# External Secrets Operator (ESO) IRSA Configurations
# ==============================================================================

# Trust relationship for the ESO service account (external-secrets:external-secrets)
data "aws_iam_policy_document" "eso_irsa_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_provider}:sub"
      values   = ["system:serviceaccount:external-secrets:external-secrets"]
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

# The IAM Role for ESO
resource "aws_iam_role" "eso_role" {
  name               = "prod-external-secrets-operator-irsa-role"
  assume_role_policy = data.aws_iam_policy_document.eso_irsa_trust.json
}

# Attach the secrets reading policy to the ESO Role
resource "aws_iam_role_policy_attachment" "attach_eso_secrets_policy" {
  role       = aws_iam_role.eso_role.name
  policy_arn = data.aws_iam_policy.existing_secrets_policy.arn
}
