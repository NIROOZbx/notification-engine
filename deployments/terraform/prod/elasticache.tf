
resource "aws_elasticache_subnet_group" "redis_subnets" {
  name       = "prod-redis-subnet-group"
  subnet_ids = module.vpc.intra_subnets
}

resource "aws_security_group" "redis_sg" {
  name        = "prod-redis-security-group"
  description = "Allow inbound TCP traffic from EKS nodes strictly"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Allow Redis access from EKS nodes"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "notification-engine-redis"
  engine               = "redis"
  engine_version       = "7.1"
  node_type            = "cache.m7g.large"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_subnets.name
  security_group_ids   = [aws_security_group.redis_sg.id]

  tags = {
    Environment = "production"
  }
}
