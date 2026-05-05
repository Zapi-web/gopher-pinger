resource "aws_security_group" "lb-sg" {
  name        = "${var.app_name}-${var.environment}-lb-sg"
  description = "security group of the load balancer"
  vpc_id      = var.vpc_id

  tags = {
    Name        = "${var.app_name}-${var.environment}-lb-sg"
    Environment = var.environment
  }
}

resource "aws_security_group_rule" "allow_http_to_lb" {
  type = "ingress"
  from_port = 80
  to_port = 80
  protocol = "tcp"
  security_group_id = aws_security_group.lb-sg.id
  cidr_blocks = ["0.0.0.0/0"]
  description = "Allow HTTP"
}

resource "aws_security_group_rule" "allow_https_to_lb" {
  type = "ingress"
  from_port = 443
  to_port = 443
  protocol = "tcp"
  security_group_id = aws_security_group.lb-sg.id
  cidr_blocks = ["0.0.0.0/0"]
  description = "Allow HTTPS"
}