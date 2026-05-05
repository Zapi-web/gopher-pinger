output "monitoring_security_group_id" {
  value       = aws_security_group.obs-server-sg.id
  description = "Monitoring security group id"
}