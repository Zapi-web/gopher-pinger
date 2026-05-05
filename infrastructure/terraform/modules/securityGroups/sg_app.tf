resource "aws_security_group" "app-server-sg" {
  name        = "${var.app_name}-${var.environment}-app-server-sg"
  description = "app server security group"
  vpc_id      = var.vpc_id

  tags = {
    Name        = "${var.app_name}-${var.environment}-app-server-sg"
    Environment = var.environment
  }
}

resource "aws_security_group_rule" "allow_lb_to_app" {
  type                     = "ingress"
  from_port                = var.app_port
  to_port                  = var.app_port
  protocol                 = "tcp"
  security_group_id        = aws_security_group.app-server-sg.id
  source_security_group_id = aws_security_group.lb-sg.id
  description              = "Allow trafic from lb"
}

resource "aws_security_group_rule" "allow_obs_to_app" {
  type                     = "ingress"
  from_port                = 9100
  to_port                  = 9100
  protocol                 = "tcp"
  security_group_id        = aws_security_group.app-server-sg.id
  source_security_group_id = aws_security_group.obs-server-sg.id
  description              = "Allow Prometheus to scrape Node Exporter"
}