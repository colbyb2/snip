terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "snip-terraform-state-982597091054"
    key            = "snip/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "snip-terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}

module "dynamodb" {
  source = "./modules/dynamodb"

  app_name    = var.app_name
  environment = var.environment
}

module "lambda" {
  source = "./modules/lambda"

  app_name            = var.app_name
  environment         = var.environment
  lambda_zip_path     = var.lambda_zip_path
  dynamodb_table_name = module.dynamodb.table_name
  dynamodb_table_arn  = module.dynamodb.table_arn
  base_url            = var.base_url
  log_level           = var.log_level
}

module "api_gateway" {
  source = "./modules/api_gateway"

  app_name             = var.app_name
  environment          = var.environment
  lambda_function_name = module.lambda.lambda_function_name
  lambda_invoke_arn    = module.lambda.invoke_arn
}
