resource "aws_instance" "app-linux-server" {
  for_each = var.subnet_ids

  ami                    = data.aws_ami.debian[var.debian_version].id
  instance_type          = var.linux_instance_type
  subnet_id              = aws_subnet.public_subnet_id[each.key].id
  vpc_security_group_ids = [aws_security_group.inst-sg.id]
  key_name               = var.key_name

  tags = {
    Name        = "${var.app_name}-${var.environment}-app-ec2-instance"
    Environment = var.environment
  }
}