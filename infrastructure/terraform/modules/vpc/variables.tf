variable "app_name" {
  type        = string
  default     = "default-app"
  description = "App name"
}

variable "environment" {
  type        = string
  default     = "test"
  description = "App environment"

  validation {
    condition     = contains(["dev", "test", "prod"], var.environment)
    error_message = "Environment must be dev,test or prod"
  }
}

variable "subnet_config" {
  description = "config for subnets"

  type = map(object({
    cidr_block = string
    az         = string
    type       = string
  }))

  default = {
    "app_1" = { cidr_block = "10.0.1.0/24", az = "eu-central-1a", type = "app" }
    "app_2" = { cidr_block = "10.0.2.0/24", az = "eu-central-1b", type = "app" }
    "db"    = { cidr_block = "10.0.3.0/24", az = "eu-central-1a", type = "db" }
    "obs"   = { cidr_block = "10.0.4.0/24", az = "eu-central-1a", type = "obs" }
  }
  validation {
    condition     = contain(["app", "db", "obs"], var.subnet_config.type)
    error_message = "type must be app, db or obs only"
  }
}