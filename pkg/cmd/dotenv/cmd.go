package dotenv

import (
	"fmt"
	"log"
	"os"

	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Cmd ro write the credentials to a dotenv file.
type Cmd struct {
	cmd.RootCommand
	Filename string
}

// NewCmd creates a command to write the credentials to a dotenv file.
func NewCmd() *cobra.Command {
	c := Cmd{
		RootCommand: cmd.RootCommand{
			Command: cobra.Command{
				Use:   "dotenv",
				Short: "stores the credentials as a dotenv file",
				Long: `
Stores the credentials as the dotenv file .gitlab-aws-credentials.env, which allow the AWS credentials to be 
shared between Gitlab jobs through the environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
and AWS_SESSION_TOKEN.

The dotenv filename defaults to .gitlab-aws-credentials.env but can be overridden through the environment
variable GITLAB_AWS_DOTENV_FILE or the command line option -f.

The following gitlab-ci.yml snippets shows the usage of the dotenv command:

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
		- gitlab-aws-credential-helper dotenv
	  artifacts:
		expire_in: 1 hour
		reports:
		  dotenv: .gitlab-aws-credentials.env
	
	dotenv:
	  stage: build
	  image:
		name: public.ecr.aws/aws-cli/aws-cli:2.13.17
		entrypoint: [""]
	  script:
		- aws sts get-caller-identity
	  needs:
		- get-aws-credentials

Note that the dotenv file with the credentials will be available for download from the pipeline artifacts.
`,
			},
		},
	}

	c.AddPersistentFlags()
	if c.Filename = os.Getenv("GITLAB_AWS_DOTENV_FILE"); c.Filename == "" {
		c.Filename = ".gitlab-aws-credentials.env"
	}
	c.Flags().StringVarP(&c.Filename, "filename", "f", c.Filename, "the name of the dotenv file")

	c.RunE = func(cmd *cobra.Command, args []string) error {
		return WriteDotEnv(c.Filename, c.Credentials)
	}

	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		if c.Filename == "" {
			return errors.New("no --filename was specified or GITLAB_AWS_DOTENV_FILE was empty.")
		}
		return nil
	}

	return &c.Command
}

func WriteDotEnv(filename string, credentials *awssts.Credentials) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file %s", filename)
		}
	}()

	writeEnvVar := func(key string, value *string) {
		var empty string
		if value == nil {
			value = &empty
		}
		line := fmt.Sprintf("%s=%s\n", key, *value)
		_, err = file.WriteString(line)
		if err != nil {
			log.Fatalf("error writing environment variable value to file, %s", err)
		}
	}

	writeEnvVar("AWS_ACCESS_KEY_ID", credentials.AccessKeyId)
	writeEnvVar("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey)
	writeEnvVar("AWS_SESSION_TOKEN", credentials.SessionToken)

	return nil
}
