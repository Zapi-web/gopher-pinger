variable "app_name" {
  type        = string
  default = "default-app"
  description = "App name"
}

variable "environment" {
  type        = string
  default = "test"
  description = "App environment"

  validation {
    condition     = contains(["dev", "test", "prod"], var.environment)
    error_message = "Environment must be dev,test or prod"
  }
}

variable "linux_instance_type" {
  type        = string
  description = "Instance type for app-server"
  default     = "t3.micro"
}

variable "debian_version_data_id" {
  type        = string
  description = "Data id of debian version server"
}

variable "vpc_id" {
  type        = string
  description = "ID of VPC"
}

variable "subnet_id" {
  type        = string
  description = "ID of observability subnet"
}

variable "obs_security_group_id" {
  type = string
  description = "Monitoring security group id"
}

variable "key_name" {
  type        = string
  description = "SSH key"
}