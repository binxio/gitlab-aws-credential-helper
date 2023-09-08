package main

import (
	"os"

	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/awsprofile"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/dotenv"
	"github.com/binxio/gitlab-aws-credential-helper/pkg/cmd/process"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitlab-aws-credential-helper",
		Short: "get AWS access credentials based upon the Gitlab id token",
	}

	rootCmd.AddCommand(dotenv.NewCmd())
	rootCmd.AddCommand(awsprofile.NewCmd())
	rootCmd.AddCommand(process.NewCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
