package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use: "deploy",
	// Aliases: []string{"rev"},
	Short: "Deploys ctfsmd to the local host.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Deploying ctfsmd.")
	},
}
