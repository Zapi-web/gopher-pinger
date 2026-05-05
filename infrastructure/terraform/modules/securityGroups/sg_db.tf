resource "aws_security_group" "db-server-sg" {
  name   = "${var.app_name}-${var.environment}-db-server-sg"
  vpc_id = var.vpc_id

  tags = {
    Name        = "${var.app_name}-${var.environment}-db-server-sg"
    Environment = var.environment
  }
}

resource "aws_security_group_rule" "allow_app_to_db" {
  type                     = "ingress"
  from_port                = var.database_port
  to_port                  = var.database_port
  protocol                 = "tcp"
  security_group_id        = aws_security_group.db-server-sg.id
  source_security_group_id = aws_security_group.app-server-sg.id
  description              = "Allow traffic to database port from app"
}

resource "aws_security_group_rule" "allow_obs_to_db" {
  type                     = "ingress"
  from_port                = 9100
  to_port                  = 9100
  protocol                 = "tcp"
  security_group_id        = aws_security_group.db-server-sg.id
  source_security_group_id = aws_security_group.obs-server-sg.id
  description              = "Allow Prometheus to scrape Node Exporter"
}