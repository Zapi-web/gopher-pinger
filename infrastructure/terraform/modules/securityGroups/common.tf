locals {
  target_sgs = {
    "app" = aws_security_group.app-server-sg.id
    "db" = aws_security_group.db-server-sg.id
    "obs" = aws_security_group.obs-server-sg.id
  }
}

resource "aws_security_group_rule" "allow_ssh" {
  for_each = local.target_sgs

  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = [var.admin_ip]
  security_group_id = each.value
}

resource "aws_security_group_rule" "allow_egress" {
  for_each          = local.target_sgs
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = each.value
}