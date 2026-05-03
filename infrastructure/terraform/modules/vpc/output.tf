output "vpc_id" {
  description = "ID of VPC"
  value       = aws_vpc.vpc.id
}

output "public_subnet_ids" {
  description = "List of IDs of public subnets"
  value       = [for subnet in aws_aws_subnet.public-subnet : subnet.id]
}