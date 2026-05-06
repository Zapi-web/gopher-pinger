output "vpc_id" {
  description = "ID of VPC"
  value       = aws_vpc.vpc.id
}

output "app_public_subnet_ids" {
  description = "Map of IDs of public subnets"
  value = {
    for k, s in aws_subnet.public-subnet : k => s.id
    if s.tags.Layer == "app"
  }
}

output "db_public_subnet_ids" {
  description = "Map of IDs of public subnets"
  value = {
    for k, s in aws_subnet.public-subnet : k => s.id
    if s.tags.Layer == "db"
  }
}

output "obs_public_subnet_ids" {
  description = "Map of IDs of public subnets"
  value = {
    for k, s in aws_subnet.public-subnet : k => s.id
    if s.tags.Layer == "obs"
  }
}