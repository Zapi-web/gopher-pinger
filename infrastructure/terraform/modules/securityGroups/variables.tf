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

variable "app_port" {
  type        = number
  description = "Port of app"
  default     = 8080
}

variable "database_port" {
  type        = number
  description = "Port of database"
  default     = 6379
}

variable "admin_ip" {
  type        = string
  description = "Admin ip for SSH and Grafana"
}

variable "vpc_id" {
  description = "ID of VPC"
  type        = string
}