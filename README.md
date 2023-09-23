Gitlab credential helper for AWS
=================================
The goal of the utility is to make it as easy as possible, to access AWS from a Gitlab CI/CD pipeline using the gitlab pipeline [id token](https://docs.gitlab.com/ee/ci/secrets/id_token_authentication.html).
You only need to specify the AWS account number and add the Gitlab ID token: it will use the pipeline id and the 
gitlab project path slug to determine the IAM role- and session name. 

For instance, if your project path is  "binxio/aws-credential-helper-demo", the IAM role it wants to assume is
"gitlab-binxio-aws-credential-helper-demo". The ID token is expected to be in the environment
variable GITLAB_AWS_IDENTITY_TOKEN.

## usage
There are three ways to use this utility:
```
gitlab-aws-credential-helper aws-profile [flags]
gitlab-aws-credential-helper process
gitlab-aws-credential-helper env [flags]
```

- [aws-profile](#aws-profile) - updates the credentials in shared credentials in ~/.aws/credentials
- [process](#credential-process) - implements the AWS [external credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html) process interface
- [env](#env) - prints the environment variables containing the AWS credentials


## Flags
The following flags can be applied to override the sensible defaults:
```text
-A, --aws-account string               required - AWS account id to assume to role in (default $GITLAB_AWS_ACCOUNT_ID)
-r, --role-name string                 required - Name of the role to assume (default gitlab-$CI_PROJECT_PATH_SLUG)
-n, --role-session-name string         required - the role session name to use (default <role name>-$CI_PIPELINE_ID)
-j, --web-identity-token-name string   required - of the environment variable with the JWT id token (default "GITLAB_AWS_IDENTITY_TOKEN")
-d, --duration-seconds int             of the session (default 3600)
```

## AWS profile
Stores the credentials in the AWS shared credentials file under the profile name "default".

The profile name defaults to "default"  but can be overridden through the environment
variable GITLAB_AWS_PROFILE or the command line option --name/-p.

### Flags
In addition to the global flags, the following flags can be applied to override the sensible defaults:
```text
-p, --name string                      the name of AWS profile (default "default")
```

| name        | default value | override         |
|-------------|---------------|------------------|
| aws-profile | default       | --aws-profile/-p |

### aws-profile example
The following gitlab-ci.yml snippets shows the usage of the aws-profile command:

```yaml
# extract the binary as artifact into the workspace
get-credential-helper:
  stage: .pre
  image:
    name: ghcr.io/binxio/gitlab-aws-credential-helper:0.0.0-6-gc168a6d
    entrypoint: [""]
  script:
    - cp /usr/local/bin/gitlab-aws-credential-helper .
  artifacts:
    expire_in: 1 hour
    paths:
      - gitlab-aws-credential-helper

aws-profile-demo:
  stage: build
  image:
    name: public.ecr.aws/aws-cli/aws-cli:2.13.17
    entrypoint: [""]
  id_tokens:
    GITLAB_AWS_IDENTITY_TOKEN:
      aud: https://gitlab.com
  script:
    # use the credential helper
    - ./gitlab-aws-credential-helper aws-profile
    - aws sts get-caller-identity
  needs:
    - get-credential-helper
```

## Credential process
Returns the credentials on stdout as specified by the credential_process interface. The process is called
by the AWS library whenever credentials are required for access.

### Usage
`gitlab-aws-credential-helper process [flags]`

### Flags
There are no flags in addition to the global flags for the credential process helper.

## credential_process example
The following gitlab-ci.yml snippets shows the usage of the process command:

```yaml
# extract the binary as artifact into the workspace
get-credential-helper:
  stage: .pre
  image:
    name: ghcr.io/binxio/gitlab-aws-credential-helper:0.0.0-6-gc168a6d
    entrypoint: [""]
  script:
    - cp /usr/local/bin/gitlab-aws-credential-helper .
  artifacts:
    expire_in: 1 hour
    paths:
      - gitlab-aws-credential-helper

aws-profile-demo:
  stage: build
  image:
    name: public.ecr.aws/aws-cli/aws-cli:2.13.17
    entrypoint: [""]
  id_tokens:
    GITLAB_AWS_IDENTITY_TOKEN:
      aud: https://gitlab.com
  script:
    # use the credential helper
    - aws configure set credential_process "$PWD/gitlab-aws-credential-helper process"
    - aws sts get-caller-identity
  needs:
    - get-credential-helper
```

## env
Returns the credentials as the environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
and AWS_SESSION_TOKEN.  When you pass a command to execute on the command line, the command
will be executed without writing the credentials.

The following gitlab-ci.yml snippets shows the usage of the env command:

### Flags
In addition to the global flags, the following flags can be applied to override the sensible defaults:
```text
-f, --filename string                  the name of the dotenv file (default stdout)
-e, --export                           prefix the environment variables with "export " (default false)
```

### dotenv example
The following gitlab-ci.yml snippets shows the usage of the dotenv command:
```yaml
variables:
  GITLAB_AWS_ACCOUNT_ID: 123456789012

get-aws-credentials:
  stage: .pre
  id_tokens:
    GITLAB_AWS_IDENTITY_TOKEN:
      aud: https://gitlab.com
  image:
    name: ghcr.io/binxio/gitlab-aws-credential-helper:0.1.0
    entrypoint: [""]
  script:
    - gitlab-aws-credential-helper env > .gitlab-as-credentials.env
  artifacts:
    expire_in: 10 min
    reports:
      env: .gitlab-aws-credentials.env
    # Note that the env file with the credentials will be available for download from the
    # pipeline artifacts by all roles associated with the project, including guest (!).
    # See https://docs.gitlab.com/ee/user/permissions.html#gitlab-cicd-permissions

env:
  stage: build
  image:
    name: public.ecr.aws/aws-cli/aws-cli:2.13.17
    entrypoint: [""]
  script:
    - aws sts get-caller-identity
  needs:
    - get-aws-credentials
```
Note that the dotenv file with the credentials will be available for download from the pipeline artifacts
by all roles associated with the project, including guest (!).

