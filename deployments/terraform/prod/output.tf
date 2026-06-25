output "redis_endpoint" {
  description = "The internal connection address for your Redis cluster"
  value       = aws_elasticache_cluster.redis.cache_nodes[0].address
}

output "eks_cluster_name" {
  description = "The name of your EKS cluster"
  value       = module.eks.cluster_name
}

output "backend_ecr_url" {
  description = "The URL of the envoy-backend ECR repository"
  value       = aws_ecr_repository.backend.repository_url
}

output "billing_ecr_url" {
  description = "The URL of the billing-service ECR repository"
  value       = aws_ecr_repository.billing.repository_url
}


output "eso_irsa_role_arn" {
  description = "The IAM Role ARN for the External Secrets Operator"
  value       = aws_iam_role.eso_role.arn
}

output "warpstream_irsa_role_arn" {
  description = "The IAM Role ARN for the WarpStream ServiceAccount"
  value       = aws_iam_role.warpstream_role.arn
}