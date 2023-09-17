package env

import (
	"errors"
	"fmt"
	"log"
	"os"

	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd"
	"github.com/spf13/cobra"
)

// Cmd ro write the credentials to a env file.
type Cmd struct {
	cmd.RootCommand
	Filename string
	Export   bool
}

// NewCmd creates a command to write the credentials to a env file.
func NewCmd() *cobra.Command {
	c := Cmd{
		RootCommand: cmd.RootCommand{
			Command: cobra.Command{
				Use:   "env",
				Short: "returns the credentials as environment variables",
				Long: `
Returns the credentials as the environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
and AWS_SESSION_TOKEN. When you pass a command to execute on the command line, the command
will be executed without writing the credentials.

The following gitlab-ci.yml snippets shows the usage of the env command:

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
`,
			},
		},
	}

	c.AddPersistentFlags()
	c.Flags().StringVarP(&c.Filename, "filename", "f", "", "the name of the env file")
	c.Flags().BoolVarP(&c.Export, "export", "e", false, "prefix variables with export keyword")

	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		if c.Filename != "" && len(args) > 0 {
			return errors.New("either specify an output file or a command to execute")
		}
		return nil
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return ExecProcess(args, c.Credentials)
		} else {
			return WriteDotEnv(c.Filename, c.Export, c.Credentials)
		}
	}

	return &c.Command
}

func WriteDotEnv(filename string, export bool, credentials *awssts.Credentials) error {
	var err error

	file := os.Stdout
	if filename != "" {
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("failed to close file %s", filename)
			}
		}()
	}

	writeEnvVar := func(key string, value *string) {
		var empty, prefix string
		if value == nil {
			value = &empty
		}
		if export {
			prefix = "export "
		}
		line := fmt.Sprintf("%s%s=%s\n", prefix, key, *value)
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
