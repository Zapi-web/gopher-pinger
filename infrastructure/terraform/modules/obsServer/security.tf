resource "aws_security_group" "obs-server-sg" {
  name = "${var.app_name}-${var.environment}-obs-server-sg"
  vpc_id = var.vpc_id

  ingress {
    from_port = 3000
    to_port = 3000
    protocol = "tcp"
    cidr_blocks = [var.admin_ip]
    description = "Grafana"
  }

  ingress {
    from_port = 3100
    to_port = 3100
    protocol = "tcp"
    security_groups = [var.app_security_group_id]
    description = "Loki"
  }

    ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = [var.admin_ip]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}