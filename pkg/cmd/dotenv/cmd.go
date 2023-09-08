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
Stores the credentials as a dotenv, to allow the AWS credentials to be shared between Gitlab jobs.

Usage:

	get_aws_credentials:
	  image: ghcr.io/binxio/gitlab-aws-credential-helper
	  id_tokens:
		WEB_IDENTITY_TOKEN:
		  aud: https://gitlab.com
	  variables:
		AWS_ACCOUNT_ID: 123234344352
      artifacts:
        reports:
          dotenv: .gitlab-aws-credentials.env
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
		line := fmt.Sprintf("%s=\"%s\"\n", key, *value)
		file.WriteString(line)
	}

	writeEnvVar("AWS_ACCESS_KEY_ID", credentials.AccessKeyId)
	writeEnvVar("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey)
	writeEnvVar("AWS_SESSION_TOKEN", credentials.SessionToken)

	return nil
}
