package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: fmt.Sprintf("Runs %s in the local host.", executableName),
	// Long: "Runs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run ctfsmd.")
	},
}
