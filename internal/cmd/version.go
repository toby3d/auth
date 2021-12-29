package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

//nolint: gochecknoglobals
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Long:  "prints the build information",
	Run:   printVersion,
}

//nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion(cmd *cobra.Command, args []string) {
	fmt.Println("IndieAuth version", runtime.Version(), runtime.GOOS) //nolint: forbidigo
}
