provider "aws" {
  region = var.aws_region
}

# ECS Cluster
resource "aws_ecs_cluster" "this" {
  name = var.ecs_cluster_name
}

# IAM Role for ECS task execution (to allow ECS Tasks to pull images and send logs)
resource "aws_iam_role" "ecs_task_exec_role" {
  name = "${var.app_name}-task-exec-role"
  assume_role_policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }]
  })
}

# Attach the AWS managed policy for ECS task execution to the role (for ECR and CloudWatch Logs access)
resource "aws_iam_role_policy_attachment" "ecs_task_exec_role_policy" {
  role       = aws_iam_role.ecs_task_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "ecs_task_exec_role_policy_ssm" {
  role       = aws_iam_role.ecs_task_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# CloudWatch Log Group for ECS task logs
resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/${var.app_name}"
  retention_in_days = 7
}

# ECS Task Definition
resource "aws_ecs_task_definition" "this" {
  family                   = var.app_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_exec_role.arn
  task_role_arn            = aws_iam_role.ecs_task_exec_role.arn
  container_definitions    = jsonencode([
    {
      name      = var.app_name
      image     = var.container_image
      essential = true
      portMappings = [
        {
          containerPort = var.container_port
          hostPort      = var.container_port
          protocol      = "tcp"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs_logs.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = var.app_name
        }
      }
    }
  ])
}

data "aws_vpc" "self" {
  filter {
    name   = "tag:Name"
    values = ["blockchain-vpc"]
  }
}

data "aws_subnets" "public" {
  filter {
    name   = "tag:Name"
    values = ["*blockchain-subnet-public*"]
  }
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.self.id]
  }
}

data "aws_security_groups" "self" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.self.id]
  }
}

# ECS Service (running on Fargate)
resource "aws_ecs_service" "this" {
  name            = "blockchain-client-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.public.ids
    security_groups  = data.aws_security_groups.self.ids
    assign_public_ip = true
  }

  health_check_grace_period_seconds = 60
  enable_execute_command = true

  deployment_controller {
    type = "ECS"
  }

  depends_on = [
    aws_iam_role_policy_attachment.ecs_task_exec_role_policy,
    aws_iam_role_policy_attachment.ecs_task_exec_role_policy_ssm,
  ]
}