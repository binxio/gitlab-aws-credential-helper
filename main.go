package main

import (
	"os"

	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/awsprofile"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/dotenv"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/process"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitlab-aws-credential-helper",
		Short: "get AWS access credentials based upon the Gitlab id token",
		Long: `
Get the AWS credentials using the gitlab pipeline id token. To make this as easy as possible, it will use
the pipeline id and the gitlab project path slug to determine the role and session name.  Just add the AWS 
account number and the Gitlab ID token. 

If your project is called "binxio-aws-credential-helper-demo", the IAM role it wants to assume is
"gitlab-binxio-aws-credential-helper-demo". The ID token is expected to be in the environment
variable GITLAB_AWS_IDENTITY_TOKEN.

The following table shows the default values for the call:

| name                    | default value                   | override                     |
+-------------------------+---------------------------------+------------------------------+
| role name               | gitlab-$CI_PROJECT_PATH_SLUG    | --role-name/-r               |
| role session name       | <role name>-$CI_PIPELINE_ID     | --role-session-name/-n       |
| aws account id          | $GITLAB_AWS_ACCOUNT_ID          | --aws-account/-A             |
| duration seconds        | $GITLAB_AWS_DURATION_SECONDS    | --duration-seconds/-d        |
| web identity token name | GITLAB_AWS_IDENTITY_TOKEN       | --web-identity-token-name/-j |

The following gitlab-ci.yml snippets shows the minimal configuration of the credential helper:

	get-aws-credentials:
	  stage: .pre
      variable:
        GITLAB_AWS_ACCOUNT_ID: 123456789012
	  id_tokens:
		GITLAB_AWS_IDENTITY_TOKEN:
		  aud: https://gitlab.com
	  image:
		name: ghcr.io/binxio/gitlab-aws-credential-helper:0.1.0
		entrypoint: [""]
	  script:
		- gitlab-aws-credential-helper dotenv
	  artifacts:
		reports:
		  dotenv: .gitlab-aws-credentials.env
`,
	}
	rootCmd.AddCommand(dotenv.NewCmd())
	rootCmd.AddCommand(awsprofile.NewCmd())
	rootCmd.AddCommand(process.NewCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
