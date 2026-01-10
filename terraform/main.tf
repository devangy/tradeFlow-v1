provider "aws" {
  region = "ap-south-1"
}

# generate the RSA private key
resource "tls_private_key" "main_key" {
algorithm = "RSA"
rsa_bits  = 4096
}

# register the public ey with AWS
resource "aws_key_pair" "deployer" {
key_name   = "my-terraform-key"
public_key = tls_private_key.main_key.public_key_openssh
}

# save the private key to  local machine
resource "local_file" "ssh_key" {
filename        = "${pathexpand("~/.ssh/aws-key.pem")}"
content         = tls_private_key.main_key.private_key_pem
file_permission = "0400" # Important: Sets read-only permission for security
}



// use default vpc alread created
data "aws_vpc" "default" {
  default = true
}


// defining our security group here for allowed traffic on our server defined ports
resource "aws_security_group" "bot_firewall" {
  name        = "bot-firewall"
  description = "Allow SSH and Web traffic"
  vpc_id      = data.aws_vpc.default.id
}

// open ssh port
resource "aws_vpc_security_group_ingress_rule" "allow_ssh" {
  security_group_id = aws_security_group.bot_firewall.id

  # ssh allowed only from my ip
  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 22
  ip_protocol = "tcp"
  to_port     = 22
}

// inbound rule for all ipv4
resource "aws_vpc_security_group_ingress_rule" "allow_http" {
  security_group_id = aws_security_group.bot_firewall.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  ip_protocol       = "tcp"
  to_port           = 80
}

# outbound Rule allow all outbound traffic
resource "aws_vpc_security_group_egress_rule" "allow_all_outbound" {
  security_group_id = aws_security_group.bot_firewall.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

// new aws instance
resource "aws_instance" "bot_instance" {
  ami           = "ami-02b8269d5e85954ef" // amazon ami image id (ubuntu 64 bit architecture)
  instance_type = "t3.micro"
  # user_data     = file("../scripts/startup.sh")
  user_data = templatefile("${path.module}/scripts/startup.sh", {
    runner_token = var.gh_runner_token
  })
  key_name = aws_key_pair.deployer.key_name
  vpc_security_group_ids = [aws_security_group.bot_firewall.id]

  // dashboard instance name
  tags = {
    Name = "tg-bot"
  }
}
