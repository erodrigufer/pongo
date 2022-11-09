package cli

import (
	"fmt"

	semver "github.com/erodrigufer/go-semver"
	"github.com/spf13/cobra"
)

var revisionCmd = &cobra.Command{
	Use:     "revision",
	Aliases: []string{"rev"},
	Short:   "Show the VCS revision hash with which the binary was built.",
	Long:    "Print the VCS revision with which the binary was built, if the binary is a modified version of a commit the suffix '-dirty' will be added to the revision hash. If the binary was not built from a repository or the commit hash could not be retrieved 'unavailable' will be printed.",
	Run: func(cmd *cobra.Command, args []string) {
		buildRev, _ := semver.GetRevision()
		fmt.Printf("%s revision: %s\n", executableName, buildRev)
	},
}

// configureRevisionCmd, sets all necessary flags and adds the revision command,
// as a child command of root command.
func configureRevisionCmd(parentCmd *cobra.Command) {
	parentCmd.AddCommand(revisionCmd)
}
