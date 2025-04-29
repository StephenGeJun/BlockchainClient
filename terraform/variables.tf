variable "aws_region" {
  description = "AWS region to deploy resources in"
  type        = string
  default     = "eu-west-1"
}

variable "aws_account_id" {
  description = "AWS Account ID (for ECR image URLs, if needed)"
  type        = string
}

variable "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
  default     = "blockchain"
}

variable "app_name" {
  description = "Application name (used for various resource naming)"
  type        = string
  default     = "blockchain-client"
}

variable "container_image" {
  description = "Container image URI for the blockchain client"
  type        = string
}

variable "container_port" {
  description = "Container port on which the application listens"
  type        = number
  default     = 8080
}