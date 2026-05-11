resource "aws_security_group" "obs-server-sg" {
  name   = "${var.app_name}-${var.environment}-obs-server-sg"
  vpc_id = var.vpc_id

  tags = {
    Name        = "${var.app_name}-${var.environment}-obs-server-sg"
    Environment = var.environment
  }
}

resource "aws_security_group_rule" "allow_grafana" {
  type              = "ingress"
  from_port         = 3000
  to_port           = 3000
  protocol          = "tcp"
  cidr_blocks       = [var.admin_ip]
  security_group_id = aws_security_group.obs-server-sg.id
  description       = "Grafana"
}

resource "aws_security_group_rule" "allow_loki" {
  type      = "ingress"
  from_port = 3100
  to_port   = 3100
  protocol  = "tcp"

  security_group_id        = aws_security_group.obs-server-sg.id
  source_security_group_id = aws_security_group.app-server-sg.id
  description              = "Loki"
}