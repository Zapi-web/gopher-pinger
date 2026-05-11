resource "aws_lb" "load_balancer" {
  name               = "${var.app_name}-${var.environment}-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.lb-sg-id]
  subnets            = values(var.subnet_ids)

  tags = {
    Name        = "${var.app_name}-${var.environment}-lb"
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "load_balacer_trg" {
  name     = "${var.app_name}-${var.environment}-lb-trg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id


  health_check {
    path                = "/health"
    interval            = 30
    healthy_threshold   = 2
    unhealthy_threshold = 5
  }
  tags = {
    Name        = "${var.app_name}-${var.environment}-lb-trg"
    Environment = var.environment
  }
}

resource "aws_lb_listener" "lb-listener" {
  load_balancer_arn = aws_lb.load_balancer.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.load_balacer_trg.id
  }
}

resource "aws_lb_target_group_attachment" "lb_attachment" {
  for_each = aws_instance.app-linux-server

  target_group_arn = aws_lb_target_group.load_balacer_trg.arn
  target_id        = each.value.id
  port             = 80
}