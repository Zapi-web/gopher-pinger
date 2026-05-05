output "app_sg_id" {
  value       = aws_security_group.app-server-sg.id
  description = "app security group id"
}

output "db_sg_id" {
  value       = aws_security_group.db-server-sg.id
  description = "database securirity group id"
}

output "obs_sg_id" {
  value       = aws_security_group.obs-server-sg.id
  description = "monitoring security group id"
}

output "lb_sg_id" {
  value       = aws_security_group.lb-sg.id
  description = "load balancer security group id"
}