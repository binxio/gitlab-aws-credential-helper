variable gitlab_token {
  type = string
  description = "the gitlab access token"
}


provider "gitlab" {
  token = var.gitlab_token
}

resource "gitlab_project" "demo" {
  name = "aws-credential-helper-demo"
}

data "external" "thumbprint" {
  program = ["./get-thumbprint"]
}


resource "aws_iam_openid_connect_provider" "gitlab" {
  url             = "https://gitlab.com"
  client_id_list  = ["https://gitlab.com"]
  thumbprint_list = [data.external.thumbprint.result.value]
}


resource "aws_iam_role" "gitlab_pipeline" {
  name = local.role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Federated = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/gitlab.com"
        },
        Action = "sts:AssumeRoleWithWebIdentity",
        Condition = {
          StringLike = {
            "gitlab.com:sub" = "project_path:${gitlab_project.demo.path_with_namespace}:ref_type:branch:ref:*"
          }
        }
      }
    ]
  })
  inline_policy {
    name        = "MetaInformationAccess"

    policy = jsonencode({
      Description = "Policy for meta information access",
      Statement = [
        {
          Effect = "Allow",
          Action = "ecr:GetAuthorizationToken",
          Resource = "*"
        },
        {
          Effect = "Allow",
          Action = "sts:GetCallerIdentity",
          Resource = "*"
        },
        {
          Effect = "Allow",
          Action = "ec2:DescribeRegions",
          Resource = "*"
        }
      ]
    })
  }

  max_session_duration = 7200
}

locals {
  role_name = substr(format("gitlab-%s", replace(gitlab_project.demo.path_with_namespace, "/[^[A-Za-z0-9-]/", "-")), 0, 64)
}

data "aws_caller_identity" "current" {}