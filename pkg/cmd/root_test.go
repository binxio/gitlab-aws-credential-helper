package cmd

import (
	"os"
	"testing"
)

func mustSetenv(t *testing.T, name, value string) {
	if err := os.Setenv(name, value); err != nil {
		t.Fatalf("failed to setenv %s to %s, %s", name, value, err)
	}
}

func TestSetDefaultsWithDefaults(t *testing.T) {
	// Mock environment variables
	mustSetenv(t, "CI_PIPELINE_ID", "12345")

	mustSetenv(t, "CI_PROJECT_PATH_SLUG", "project_slug")
	mustSetenv(t, "GITLAB_AWS_ACCOUNT_ID", "aws_account_id")

	// Create an instance of the struct
	c := &RootCommand{}

	// Call the function being tested
	c.SetDefaults()

	// Check if the values are set correctly
	expectedPipelineId := "12345"
	if c.PipelineId != expectedPipelineId {
		t.Errorf("PipelineId is not set correctly. Expected: %s, Got: %s", expectedPipelineId, c.PipelineId)
	}

	expectedRoleName := "gitlab-project_slug"
	if c.RoleName != expectedRoleName {
		t.Errorf("RoleName is not set correctly. Expected: %s, Got: %s", expectedRoleName, c.RoleName)
	}

	expectedAwsAccount := "aws_account_id"
	if c.AwsAccount != expectedAwsAccount {
		t.Errorf("AwsAccount is not set correctly. Expected: %s, Got: %s", expectedAwsAccount, c.AwsAccount)
	}

	expectedDurationSeconds := int64(3600)
	if c.DurationSeconds != expectedDurationSeconds {
		t.Errorf("DurationSeconds is not set correctly. Expected: %d, Got: %d", expectedDurationSeconds, c.DurationSeconds)
	}

	expectedWebIdentityTokenName := "GITLAB_AWS_IDENTITY_TOKEN"
	if c.WebIdentityTokenName != expectedWebIdentityTokenName {
		t.Errorf("WebIdentityTokenName is not set correctly. Expected: %s, Got: %s", expectedWebIdentityTokenName, c.WebIdentityTokenName)
	}
}

func TestSetDefaultsWithEnvOverride(t *testing.T) {
	// Mock environment variables
	mustSetenv(t, "CI_PIPELINE_ID", "654321")

	mustSetenv(t, "CI_PROJECT_PATH_SLUG", "project_slug")
	mustSetenv(t, "GITLAB_AWS_ACCOUNT_ID", "aws_account_id")
	mustSetenv(t, "GITLAB_AWS_DURATION_SECONDS", "1800")
	mustSetenv(t, "GITLAB_AWS_IDENTITY_TOKEN_NAME", "identity_token_name")

	// Create an instance of the struct
	c := &RootCommand{}

	// Call the function being tested
	c.SetDefaults()

	// Check if the values are set correctly
	expectedPipelineId := "654321"
	if c.PipelineId != expectedPipelineId {
		t.Errorf("PipelineId is not set correctly. Expected: %s, Got: %s", expectedPipelineId, c.PipelineId)
	}

	expectedRoleName := "gitlab-project_slug"
	if c.RoleName != expectedRoleName {
		t.Errorf("RoleName is not set correctly. Expected: %s, Got: %s", expectedRoleName, c.RoleName)
	}

	expectedAwsAccount := "aws_account_id"
	if c.AwsAccount != expectedAwsAccount {
		t.Errorf("AwsAccount is not set correctly. Expected: %s, Got: %s", expectedAwsAccount, c.AwsAccount)
	}

	expectedDurationSeconds := int64(1800)
	if c.DurationSeconds != expectedDurationSeconds {
		t.Errorf("DurationSeconds is not set correctly. Expected: %d, Got: %d", expectedDurationSeconds, c.DurationSeconds)
	}

	expectedWebIdentityTokenName := "identity_token_name"
	if c.WebIdentityTokenName != expectedWebIdentityTokenName {
		t.Errorf("WebIdentityTokenName is not set correctly. Expected: %s, Got: %s", expectedWebIdentityTokenName, c.WebIdentityTokenName)
	}
}

func TestSetDefaultsInvalidDuration(t *testing.T) {
	// Mock environment variables
	mustSetenv(t, "GITLAB_AWS_DURATION_SECONDS", "invalid_duration")
	c := &RootCommand{}
	c.SetDefaults()
	if c.DurationSeconds != 3600 {
		t.Errorf("expected the default of 3600 as duration seconds")
	}
}
