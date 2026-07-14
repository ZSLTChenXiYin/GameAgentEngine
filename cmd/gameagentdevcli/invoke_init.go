package main

func init() {
	rootCmd.PersistentFlags().String("pipeline-mode", "", "Invoke context pipeline mode: vertical / polling / full")
	rootCmd.PersistentFlags().String("dynamic-interfaces-json", "", "Request-scoped dynamic_interfaces JSON array")
	rootCmd.PersistentFlags().String("dynamic-interfaces-file", "", "Read request-scoped dynamic_interfaces JSON array from file")
	rootCmd.AddCommand(invokeCmd)
}
