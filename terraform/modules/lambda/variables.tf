variable "app_name" {
  description = "Application name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "lambda_zip_path" {
  description = "Path to the Lambda deployment zip file"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table (for IAM permissions)"
  type        = string
}

variable "base_url" {
  description = "Base URL for generated short links"
  type        = string
}

variable "log_level" {
  description = "Log level for the application"
  type        = string
  default     = "info"
}
