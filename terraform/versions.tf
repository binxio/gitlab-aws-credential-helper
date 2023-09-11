terraform {
  required_providers {
    gitlab = {
      version = "~> 16.3.0"
      source  = "gitlabhq/gitlab"
    }
    aws = {
      version = "~> 5.16.0"
    }
    template = {
      version = "~>2.2.0"
    }
  }

  required_version = "~> 1.5.7"
}

