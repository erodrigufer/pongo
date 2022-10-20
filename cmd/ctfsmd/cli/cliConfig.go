package cli

func configCLI() {
	// Add a child command to the root command.
	rootCmd.AddCommand(deployCmd)
}
