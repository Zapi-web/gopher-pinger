resource "aws_instance" "app-linux-server" {
  for_each = var.subnet_ids

  ami                    = var.debian_version_data_id
  instance_type          = var.linux_instance_type
  subnet_id              = each.value
  vpc_security_group_ids = [var.app-sg-id]
  key_name               = var.key_name

  tags = {
    Name        = "${var.app_name}-${var.environment}-app-ec2-instance-${each.key}"
    Environment = var.environment
    Role        = "app"
  }
}