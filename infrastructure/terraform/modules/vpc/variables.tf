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

variable "subnet_config" {
  description = "config for subnets"

  type = map(object({
    cidr_block = string
    az         = string
    type       = string
  }))

  validation {
    condition     = contain(["app", "db", "obs"], var.subnet_config.type)
    error_message = "type must be app, db or obs only"
  }
}