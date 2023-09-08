package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	awssession "github.com/aws/aws-sdk-go/aws/session"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

// RootCommand the root command with all the global flags
type RootCommand struct {
	cobra.Command
	RoleName             string
	RoleSessionName      string
	AwsAccount           string
	DurationSeconds      int64
	PipelineId           string
	WebIdentityTokenName string
	WebIdentityToken     string
	STS                  *awssts.STS
	RoleArn              string
	Credentials          *awssts.Credentials
}

// AddPersistentFlags adds all the persistent flags to the command
func (c *RootCommand) AddPersistentFlags() {
	c.PersistentFlags().SortFlags = false
	c.SetDefaults()
	c.Flags().StringVarP(&c.RoleName, "role-name", "r", c.RoleName, "Name of the role to assume")
	c.Flags().StringVarP(&c.RoleSessionName, "role-session-name", "n", "", "the role session name to use")
	c.Flags().StringVarP(&c.AwsAccount, "aws-account", "A", c.AwsAccount, "AWS account id to assume to role in")
	c.Flags().Int64VarP(&c.DurationSeconds, "duration-seconds", "d", 3600, "of the session")
	c.Flags().StringVarP(&c.WebIdentityTokenName, "web-identity-token-name", "j", c.WebIdentityTokenName, "of the environment variable with the JWT id token")
	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return c.GetSTSCredentials()
	}
}

// SetDefaults sets the defaults for the root command.
func (c *RootCommand) SetDefaults() {
	c.PipelineId = os.Getenv("CI_PIPELINE_ID")

	if slug := os.Getenv("CI_PROJECT_PATH_SLUG"); slug != "" {
		c.RoleName = fmt.Sprintf("gitlab-%.57s", slug)
	}

	if accountId := os.Getenv("GITLAB_AWS_ACCOUNT_ID"); accountId != "" {
		c.AwsAccount = accountId
	}

	if durationSeconds := os.Getenv("GITLAB_AWS_DURATION_SECONDS"); durationSeconds != "" {
		if seconds, err := strconv.Atoi(durationSeconds); err == nil && seconds > 0 {
			c.DurationSeconds = int64(seconds)
		} else {
			log.Fatalf("the environment variable GITLAB_AWS_DURATION_SECONDS is not a positive integer")
		}
	}

	if c.WebIdentityTokenName = os.Getenv("GITLAB_AWS_IDENTITY_TOKEN_NAME"); c.WebIdentityTokenName == "" {
		c.WebIdentityTokenName = "GITLAB_AWS_IDENTITY_TOKEN"
	}
}

// GetSTSCredentials gets the STS credentials based upon the gitlab pipeline id token.
func (c *RootCommand) GetSTSCredentials() error {
	if c.RoleName == "" {
		return errors.New("the role name is not set. Perhaps the environment variable CI_PROJECT_PATH_SLUG is not present")
	}
	if len(c.RoleName) > 64 {
		return errors.New("the role name exceeds the maximum of 64 characters allowed by AWS")
	}
	if c.AwsAccount == "" {
		return errors.New("the AWS account is not set. Use --aws-account or set the environment variable GITLAB_AWS_ACCOUNT_ID")
	}

	c.RoleArn = fmt.Sprintf("arn:aws:iam::%s:role/%s", c.AwsAccount, c.RoleName)

	if c.WebIdentityToken = os.Getenv(c.WebIdentityTokenName); c.WebIdentityToken != "" {
		return errors.New(fmt.Sprintf("the environment variable %s is not set", c.WebIdentityTokenName))
	}

	if c.RoleSessionName == "" {
		if c.PipelineId == "" {
			c.RoleSessionName = c.RoleName
		} else {
			c.RoleSessionName = fmt.Sprintf("%s-%s", c.RoleName, c.PipelineId)
		}
	}

	var err error
	var session *awssession.Session
	session, err = awssession.NewSession(
		&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials("", "", ""),
		},
	)
	if err != nil {
		return err
	}

	c.STS = awssts.New(session)

	input := &awssts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(c.RoleArn),
		RoleSessionName:  aws.String(c.RoleSessionName),
		WebIdentityToken: aws.String(c.WebIdentityToken),
		DurationSeconds:  aws.Int64(c.DurationSeconds),
	}

	result, err := c.STS.AssumeRoleWithWebIdentity(input)
	if err == nil {
		c.Credentials = result.Credentials
	} else {
		return err
	}

	return nil
}
