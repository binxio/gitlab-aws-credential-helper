Gitlab credential helper for AWS
=================================
This utility allows you to easily get AWS credentials using the gitlab pipeline [id token](https://docs.gitlab.com/ee/ci/secrets/id_token_authentication.html).

There are three ways to use this utility:

```yaml
get_aws_credentials:
  variables:
    AWS_ACCOUNT_ID: 123234344352
  image: 
    name: ghcr.io/binxio/gitlab-aws-credential-helper
  id_tokens:
    GITLAB_AWS_IDENTITY_TOKEN:
      aud: https://gitlab.com
  script:
    artifacts:
      reports:
        dotenv: .gitlab-aws-credentials.env
```

The credential helper will call the [assume-role-with-web-identity](https://docs.aws.amazon.com/cli/latest/reference/sts/assume-role-with-web-identity.html) operation using the following parameters:

| parameter | value |
|-----------+--------|
| web-identity-token | value of the environment variable GITLAB_AWS_IDENTITY_TOKEN |
| role-name | `gitlab-$CI_PROJECT_PATH_SLUG` truncated to 64 characters |
