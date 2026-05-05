output "vpc_id" {
  description = "ID of VPC"
  value       = aws_vpc.vpc.id
}

output "app_public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = [for subnet in aws_subnet.public-subnet : subnet.id if lookup(subnet.tags, "Layer", "") == "app"]
}

output "db_public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = [for subnet in aws_subnet.public-subnet : subnet.id if lookup(subnet.tags, "Layer", "") == "db"]
}

output "obs_public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = [for subnet in aws_subnet.public-subnet : subnet.id if lookup(subnet.tags, "Layer", "") == "obs"]
}