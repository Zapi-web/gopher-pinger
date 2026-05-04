output "app_security_group_id" {
    value = aws_security_group.app-server-sg.id
    description = "Security group id of app"
}