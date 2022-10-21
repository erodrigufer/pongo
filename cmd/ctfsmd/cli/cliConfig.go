package cli

func configCLI() {
	configureRunCmd(rootCmd)
	rootCmd.AddCommand(revisionCmd)
}
