variable "app_instance_size" {
  type        = string
  default     = "t3.micro"
  description = "App instance type"
}

variable "database_instance_size" {
  type        = string
  default     = "t3.micro"
  description = "Database instance type"
}

variable "observability_instance_size" {
  type        = string
  default     = "t3.small"
  description = "Observability instance type"
}

variable "app_name" {
  type        = string
  default     = "default-app"
  description = "App name"
}

variable "key_name" {
  type        = string
  description = "key_name"
}

variable "admin_ip" {
  type        = string
  description = "admin ip for ssh"
}

variable "app_port" {
  type        = number
  default     = 80
  description = "App port"
}

variable "environment" {
  type        = string
  default     = "dev"
  description = "Environment"

  validation {
    condition     = contains(["dev", "test", "prod"], var.environment)
    error_message = "Environment must be dev,test or prod"
  }
}

variable "debian_version" {
  default     = "12"
  type        = string
  description = "Debian version"

  validation {
    condition     = contains(["10", "11", "12", "13"], var.debian_version)
    error_message = "Only 10-13 versions"
  }
}