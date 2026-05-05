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

variable "subnet_ids" {
  type        = list(string)
  description = "IDs of public subnets"
}

variable "app_port" {
  type        = number
  description = "Port of app"
  default     = 8080
}

variable "monitoring_security_group_id" {
  type = string
  description = "Id of monitoring security group"
}

variable "key_name" {
  type        = string
  description = "SSH key"
}