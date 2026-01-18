resource "aws_dynamodb_table" "links" {
  name         = "${var.app_name}-${var.environment}-links"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "short_code"

  attribute {
    name = "short_code"
    type = "S"
  }

  tags = {
    Name        = "${var.app_name}-${var.environment}-links"
    Environment = var.environment
    Project     = var.app_name
  }
}
