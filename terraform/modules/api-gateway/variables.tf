variable "app_name" {
  description = "Application name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "lambda_function_name" {
  description = "Name of the Lambda function to invoke"
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  type        = string
}

variable "base_url" {
  description = "Base URL for generated short links"
  type        = string
  default     = "https://es74k3z5m1.execute-api.us-east-1.amazonaws.com"
}
