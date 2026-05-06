resource "aws_lb" "load_balancer" {
  for_each = var.subnet_ids

  name               = "${var.app_name}-${var.environment}-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.lb-sg-id]
  subnets            = [each.value]

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
  for_each = var.subnet_ids

  load_balancer_arn = aws_lb.load_balancer[each.key].arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.load_balacer_trg.id
  }
}

resource "aws_lb_target_group_attachment" "lb_attachment" {
  for_each = {
    for pair in local.instance_port_matrix : "${pair[0]}-${pair[1]}" => {
      instance_key = pair[0]
      port         = pair[1]
    }
  }
  target_group_arn = aws_lb_target_group.load_balacer_trg.arn
  target_id        = aws_instance.app-linux-server[each.value.instance_key].id
  port             = each.value.port
}

locals {
  instance_port_matrix = setproduct(keys(aws_instance.app-linux-server), ["80", "443"])
}