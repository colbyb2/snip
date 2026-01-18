variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "app_name" {
  description = "Application name, used for resource naming"
  type        = string
  default     = "snip"
}

variable "lambda_zip_path" {
  description = "Path to the Lambda deployment zip file"
  type        = string
  default     = "../build/lambda.zip"
}

variable "base_url" {
  description = "Base URL for generated short links"
  type        = string
  default     = "https://es74k3z5m1.execute-api.us-east-1.amazonaws.com"
}

variable "log_level" {
  description = "Log level for the application"
  type        = string
  default     = "info"
}
