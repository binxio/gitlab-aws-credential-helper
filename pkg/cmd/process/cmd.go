package process

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials/processcreds"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd"
	"github.com/spf13/cobra"
)

// Cmd to respond to the AWS credential_process
type Cmd struct {
	cmd.RootCommand
}

// NewCmd creates a command to respond as a AWS credential process
func NewCmd() *cobra.Command {
	c := Cmd{
		RootCommand: cmd.RootCommand{
			Command: cobra.Command{
				Use:   "process",
				Short: "returns the credentials as required by the AWS credential_process",
				Long: `
Returns the credentials on stdout as specified by the credential_process interface. The process is called
by the AWS library whenever credentials are required for access.

The following gitlab-ci.yml snippets shows the usage of the process command:

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
`,
			},
		},
	}

	c.AddPersistentFlags()

	c.RunE = func(cmd *cobra.Command, args []string) error {
		return WriteProcessCredentials(c.Credentials)
	}

	return &c.Command
}

func WriteProcessCredentials(credentials *sts.Credentials) (err error) {
	var encoded []byte

	if encoded, err = json.MarshalIndent(
		processcreds.CredentialProcessResponse{
			Version:         1,
			AccessKeyID:     *credentials.AccessKeyId,
			SecretAccessKey: *credentials.SecretAccessKey,
			SessionToken:    *credentials.SessionToken,
			Expiration:      credentials.Expiration,
		},
		"", ""); err != nil {
		return err
	}

	writer := bufio.NewWriter(os.Stdout)
	if _, err = writer.Write(encoded); err != nil {
		return err
	}
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}
