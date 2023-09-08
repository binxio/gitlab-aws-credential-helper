package awsprofile

import (
	"log"
	"os"
	"path/filepath"
	"time"

	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// Cmd to write credentials to the AWS shared credentials file
type Cmd struct {
	cmd.RootCommand
	AWSProfile string
}

// NewCmd creates a command to write the AWS shared credentials file
func NewCmd() *cobra.Command {
	c := Cmd{
		RootCommand: cmd.RootCommand{
			Command: cobra.Command{
				Use:   "aws-profile",
				Short: "stores the credentials in the AWS shared credentials file",
				Long: `
Stores the credentials in the AWS shared credentials file under the specified profile name.

Usage:

	get_aws_credentials:
	  image:
        name: ghcr.io/binxio/gitlab-aws-credential-helper
        entrypoint: [""]
	  id_tokens:
		JWT_TOKEN:
		  aud: https://gitlab.com
	  variables:
		GITLAB_AWS_ACCOUNT_ID: 123234344352
        GITLAB_AWS_PROFILE: my-profile
`,
			},
		},
	}

	c.AddPersistentFlags()
	if c.AWSProfile = os.Getenv("GITLAB_AWS_PROFILE"); c.AWSProfile == "" {
		c.AWSProfile = "default"
	}
	c.Flags().StringVarP(&c.AWSProfile, "aws-profile", "p", c.AWSProfile, "the name of AWS profile")

	c.RunE = func(cmd *cobra.Command, args []string) error {
		return WriteToSharedConfig(c.AWSProfile, c.Credentials)
	}

	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		if c.AWSProfile == "" {
			return errors.New("no --aws-profile was specified or GITLAB_AWS_PROFILE was empty.")
		}
		return nil
	}

	return &c.Command
}

func WriteToSharedConfig(profileName string, credentials *awssts.Credentials) (err error) {
	var credentialFile string
	if credentialFile = os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); credentialFile == "" {
		credentialFile = os.ExpandEnv("$HOME/.aws/credentials")
	}
	cfg, err := ini.Load(credentialFile)
	if err != nil {
		return err
	}
	var section *ini.Section
	if cfg.HasSection(profileName) {
		section = cfg.Section(profileName)
	} else {
		if section, err = cfg.NewSection(profileName); err != nil {
			return err
		}
	}
	values := map[string]string{
		"aws_access_key_id":     *credentials.AccessKeyId,
		"aws_secret_access_key": *credentials.SecretAccessKey,
		"aws_session_token":     *credentials.SessionToken,
		"expiration":            credentials.Expiration.Format(time.RFC3339),
	}
	for key, value := range values {
		if section.HasKey(key) {
			section.Key(key).SetValue(value)
		} else {
			section.NewKey(key, value)
		}
	}
	directory := filepath.Dir(credentialFile)
	if _, err := os.Stat(directory); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0o750)
		if err != nil {
			return err
		}
	}

	var file *os.File
	if file, err = os.OpenFile(credentialFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o700); err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Printf("WARNING: failed to close %s", credentialFile)
		}
	}(file)

	_, err = cfg.WriteTo(file)
	return err
}
