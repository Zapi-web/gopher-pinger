resource "aws_lb" "load_balancer" {
  name               = "${var.app_name}-${var.environment}-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb-sg.id]
  subnets            = var.subnet_ids

  tags = {
    Name        = "${var.app_name}-${var.environment}-lb"
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "load_balacer_trg" {
  name     = "${var.app_name}-${var.environment}-lb-trg"
  port     = var.app_port
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  tags = {
    Name        = "${var.app_name}-${var.environment}-lb-trg"
    Environment = var.environment
  }
}

resource "aws_lb_listener" "lb-listener" {
  load_balancer_arn = aws_lb.load_balancer.arn
  for_each = {
    80 = "HTTP"
  }
  port     = each.key
  protocol = each.value

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.load_balacer_trg.id
  }
}

resource "aws_lb_target_group_attachment" "lb_attachment" {
  for_each         = [80, 443]
  target_group_arn = aws_lb_target_group.load_balacer_trg.arn
  target_id        = aws.app_linux_server.id
  port             = each.key
}