package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	c.Flags().SortFlags = false
	c.Flags().StringVarP(&c.AwsAccount, "aws-account", "A", c.AwsAccount, "AWS account id to assume to role in (default $GITLAB_AWS_ACCOUNT_ID)")
	c.Flags().StringVarP(&c.RoleName, "role-name", "r", c.RoleName, "Name of the role to assume (default gitlab-$CI_PROJECT_PATH_SLUG)")
	c.Flags().StringVarP(&c.RoleSessionName, "role-session-name", "n", "", "the role session name to use  (default <role name>-$CI_PIPELINE_ID)`")
	c.Flags().StringVarP(&c.WebIdentityTokenName, "web-identity-token-name", "j", c.WebIdentityTokenName, "of the environment variable with the JWT id token (default GITLAB_AWS_IDENTITY_TOKEN)")
	c.Flags().Int64VarP(&c.DurationSeconds, "duration-seconds", "d", c.DurationSeconds, "of the session")
	c.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if _, err := GetDurationSecondsFromEnvironment(); err != nil {
			return err
		}
		return c.GetSTSCredentials()
	}
}

// GetDurationSecondsFromEnvironment returns the integer value from GITLAB_AWS_DURATION_SECONDS or the default 3600 if it does not exist.
// an invalid integer value, return the default value and err set.
func GetDurationSecondsFromEnvironment() (seconds int64, err error) {
	seconds = 3600
	if durationSeconds := os.Getenv("GITLAB_AWS_DURATION_SECONDS"); durationSeconds != "" {
		var s int
		if s, err = strconv.Atoi(durationSeconds); err == nil && seconds > 0 {
			return int64(s), nil
		} else {
			return seconds, errors.New("the environment variable GITLAB_AWS_DURATION_SECONDS is not a positive integer")
		}
	}
	return seconds, err
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

	c.DurationSeconds, _ = GetDurationSecondsFromEnvironment()

	if c.WebIdentityTokenName = os.Getenv("GITLAB_AWS_IDENTITY_TOKEN_NAME"); c.WebIdentityTokenName == "" {
		c.WebIdentityTokenName = "GITLAB_AWS_IDENTITY_TOKEN"
	}
}

func truncate(name string, maxLength int) string {
	if len(name) < maxLength {
		return name
	}
	return string([]rune(name)[0:maxLength])
}

// GenerateRoleSessionName generates a valid role session name based on the role name and pipeline id.
func GenerateRoleSessionName(roleName, pipelineId string) string {
	maxLength := 64
	invalidCharacters := regexp.MustCompile(`[^=,.@A-Za-z0-9_]+`)
	validRoleSessionName := strings.Trim(invalidCharacters.ReplaceAllString(roleName, "-"), "-")
	if pipelineId == "" {
		return truncate(validRoleSessionName, 64)
	} else {
		maxLength = 64 - len(pipelineId) - 1
		if maxLength <= 0 {
			return truncate(pipelineId, 64)
		}
		validRoleSessionName = truncate(validRoleSessionName, maxLength)
	}
	return fmt.Sprintf("%s-%s", validRoleSessionName, pipelineId)
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

	if c.WebIdentityToken = os.Getenv(c.WebIdentityTokenName); c.WebIdentityToken == "" {
		return errors.New(fmt.Sprintf("the environment variable %s is not set", c.WebIdentityTokenName))
	}

	if c.RoleSessionName == "" {
		c.RoleSessionName = GenerateRoleSessionName(c.RoleName, c.PipelineId)
	}

	var err error
	var session *awssession.Session
	session, err = awssession.NewSession(
		&aws.Config{
			Credentials: credentials.AnonymousCredentials,
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
