resource "gitlab_project" "demo" {
  name = "aws-credential-helper-demo"
}


resource "gitlab_repository_file" "gitlab-ci" {
  project        = gitlab_project.demo.id
  file_path      = ".gitlab-ci.yml"
  branch         = "main"
  content        = base64encode(data.template_file.gitlab-ci.rendered)
  author_email   = "mvanholsteijn@xebia.com"
  author_name    = "Mark van Holsteijn"
  commit_message = format("updated with commit %s", data.external.release.result.release)
}

data "template_file" "gitlab-ci" {
  vars = {
    release        = data.external.release.result.release
    aws_account_id = data.aws_caller_identity.current.account_id
  }
  template = file("templates/gitlab-ci.yaml.template")
}


data "external" "release" {
  program = ["./bin/get-git-release"]
}
