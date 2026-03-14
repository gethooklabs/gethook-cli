package main

import (
	"fmt"
	"os"

	"github.com/gethook/gethook-cli/internal/output"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}
