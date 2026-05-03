variable "app_name" {
  type        = string
  description = "App name"
}

variable "environment" {
  type        = string
  description = "App environment"

  validation {
    condition     = contains(["dev", "test", "prod"], var.environment)
    error_message = "Environment must be dev,test or prod"
  }
}

variable "cidr_block" {
  type        = string
  description = "cidr block for vpc"
  default     = "10.0.0.0/16"
}

variable "public_subnets" {
  type        = map(string)
  description = "public subnets and cidr blocks of them"
}