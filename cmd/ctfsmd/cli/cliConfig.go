package cli

func configCLI() {
	// Add all child commands to the root command.
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(revisionCmd)
}
