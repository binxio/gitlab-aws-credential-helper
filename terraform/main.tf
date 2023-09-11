variable "gitlab_token" {
  type        = string
  description = "the gitlab access token"
}

provider "gitlab" {
  token = var.gitlab_token
}

data "aws_caller_identity" "current" {}
