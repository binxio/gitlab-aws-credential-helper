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
Stores the credentials in the AWS shared credentials file under the profile name "default". 

The profile name defaults to "default"  but can be overridden through the environment
variable GITLAB_AWS_PROFILE or the command line option -p.

The following gitlab-ci.yml snippets shows the usage of the aws-profile command:

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
	cfg, err := ini.LooseLoad(credentialFile)
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
			_, err = section.NewKey(key, value)
			if err != nil {
				log.Printf("failed to store new key %s in the config file; %s", key, err)
			}
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
