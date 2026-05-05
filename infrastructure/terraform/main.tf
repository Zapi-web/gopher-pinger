module "vpc" {
  source = "./modules/vpc"
}

module "debian" {
  source = "./modules/debian"
}

module "security_groups" {
  source = "./modules/securityGroups"

  vpc_id   = module.vpc.vpc_id
  admin_ip = "0.0.0.0/0"
}

module "monitoring_server" {
  source = "./modules/obsServer"

  vpc_id                 = module.vpc.vpc_id
  key_name               = "-"
  subnet_ids             = module.vpc.obs_public_subnet_ids
  debian_version_data_id = module.debian.debian_version_id
  obs_security_group_id  = module.security_groups.obs_sg_id
}

module "database_server" {
  source = "./modules/dbServer"

  vpc_id                 = module.vpc.vpc_id
  key_name               = "-"
  db-sg-id               = module.security_groups.db_sg_id
  debian_version_data_id = module.debian.debian_version_id
  subnet_ids             = module.vpc.db_public_subnet_ids
}

module "app_server" {
  source = "./modules/appServer"

  vpc_id                 = module.vpc.vpc_id
  lb-sg-id               = module.security_groups.lb_sg_id
  app-sg-id              = module.security_groups.app_sg_id
  subnet_ids             = module.vpc.app_public_subnet_ids
  debian_version_data_id = module.debian.debian_version_id
  key_name               = "-"
}
