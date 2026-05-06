terraform {
  backend "s3" {
    bucket         = "amzn-s3-unique-terraform-bucket-271598835315-eu-central-1-an"
    key            = "terraform-tfstate"
    region         = "eu-central-1"
    dynamodb_table = "lock-table"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.43.0"
    }
  }
}