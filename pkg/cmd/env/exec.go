package env

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

// ExecProcess executes the command with the credentials as environment variables.
func ExecProcess(cmd []string, credentials *awssts.Credentials) error {
	program, err := exec.LookPath(cmd[0])
	if err != nil {
		return errors.Errorf("could not find program %s on path, %s", cmd[0], err)
	}

	err = syscall.Exec(program, cmd, NewEnvironmentWithCredentials(os.Environ(), credentials))
	if err != nil {
		return errors.Errorf("could not exec %s, %s", program, err)
	}
	// it never gets here if the Exec is successful.
	return nil
}

func splitEnvironmentVariable(envEntry string) (string, string) {
	result := strings.SplitN(envEntry, "=", 2)
	return result[0], result[1]
}

// NewEnvironmentWithCredentials creates a new environment variable array adding the environment variables AWS_ACCESS_KEY_ID,
// AWS_SECRET_ACCESS_KEY and AWS_SESSION_TOKEN.
func NewEnvironmentWithCredentials(env []string, credentials *awssts.Credentials) []string {
	result := make([]string, 0, len(env)+3)
	credentialValues := map[string]*string{
		"AWS_ACCESS_KEY_ID":     credentials.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY": credentials.SecretAccessKey,
		"AWS_SESSION_TOKEN":     credentials.SessionToken,
	}

	for _, envEntry := range env {
		name, _ := splitEnvironmentVariable(envEntry)
		if _, ok := credentialValues[name]; !ok {
			result = append(result, envEntry)
		}
	}

	for _, name := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"} {
		value, _ := credentialValues[name]
		if value != nil {
			result = append(result, fmt.Sprintf("%s=%s", name, *value))
		} else {
			result = append(result, fmt.Sprintf("%s=", name))
		}
	}

	return result
}
