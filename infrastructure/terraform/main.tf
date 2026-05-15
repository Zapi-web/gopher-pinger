module "vpc" {
  source = "./modules/vpc"

  app_name    = var.app_name
  environment = var.environment
}

module "debian" {
  source = "./modules/debian"

  debian_version = var.debian_version
}

module "security_groups" {
  source = "./modules/securityGroups"

  app_name    = var.app_name
  app_port    = var.app_port
  vpc_id      = module.vpc.vpc_id
  admin_ip    = var.admin_ip
  environment = var.environment
}

module "monitoring_server" {
  source = "./modules/obsServer"

  app_name               = var.app_name
  environment            = var.environment
  key_name               = var.key_name
  subnet_ids             = module.vpc.obs_public_subnet_ids
  debian_version_data_id = module.debian.debian_version_id
  obs_security_group_id  = module.security_groups.obs_sg_id
  linux_instance_type    = var.observability_instance_size
}

module "database_server" {
  source = "./modules/dbServer"

  app_name               = var.app_name
  environment            = var.environment
  key_name               = var.key_name
  db-sg-id               = module.security_groups.db_sg_id
  debian_version_data_id = module.debian.debian_version_id
  subnet_ids             = module.vpc.db_public_subnet_ids
  linux_instance_type    = var.database_instance_size
}

module "app_server" {
  source = "./modules/appServer"

  app_name               = var.app_name
  environment            = var.environment
  vpc_id                 = module.vpc.vpc_id
  lb-sg-id               = module.security_groups.lb_sg_id
  app-sg-id              = module.security_groups.app_sg_id
  subnet_ids             = module.vpc.app_public_subnet_ids
  debian_version_data_id = module.debian.debian_version_id
  key_name               = var.key_name
  linux_instance_type    = var.app_instance_size
}
