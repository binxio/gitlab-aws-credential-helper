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
returns the credentials on stdout as specified by the credential_process interface.

To use, type:

  $ aws config --profile my-profile set credential_process 'gitlab-aws-credential-helper process --aws-account 1234566777'
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
