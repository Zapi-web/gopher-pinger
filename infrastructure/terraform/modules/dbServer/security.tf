resource "aws_security_group" "db-server-sg" {
    name = "${var.app_name}-${var.environment}-db-server-sg"
    vpc_id = var.vpc_id

    ingress {
        from_port = var.database_port
        to_port = var.database_port
        protocol = "tcp"
        security_groups = [var.app_security_group_id]
        description = "allow traffic to database port from app"
    }

    ingress {
        from_port = 9100
        to_port = 9100
        protocol = "tcp"
        security_groups = [var.monitoring_security_group_id]
        description = "Allow Prometheus to scrape Node Exporter"
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = -1
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "${var.app_name}-${var.environment}-db-server-sg"
      Environment = var.environment
    }
}