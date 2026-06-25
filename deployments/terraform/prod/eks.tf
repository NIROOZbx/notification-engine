module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  name               = "envoy-cluster"
  kubernetes_version = "1.35"

  endpoint_public_access = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  enable_cluster_creator_admin_permissions = true

addons = {
  vpc-cni = {
    configuration_values = jsonencode({
      env = {
        ENABLE_PREFIX_DELEGATION = "true"
        WARM_PREFIX_TARGET       = "1"
      }
    })
  }
  coredns    = {}
  kube-proxy = {}
}

  eks_managed_node_groups = {
    envoy_nodes = {
      min_size     = 1
      max_size     = 3
      desired_size = 1

      instance_types = ["t3.small"]
      capacity_type  = "ON_DEMAND"

      # Amazon Linux 2023 uses nodeadm, which requires a manual override to calculate maxPods with prefix delegation
      cloudinit_pre_nodeadm = [
        {
          content_type = "application/node.eks.aws"
          content      = <<-EOT
            ---
            apiVersion: node.eks.aws/v1alpha1
            kind: NodeConfig
            spec:
              kubelet:
                config:
                  maxPods: 110
          EOT
        }
      ]

      tags = {
        "kubernetes.io/cluster/envoy-cluster" = "owned"
      }
    }
  }

  tags = {
    Environment = "production"
    Project     = "envoy"
  }
}
