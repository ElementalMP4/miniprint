package main

import (
	"fmt"
	"miniprint/printer"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "miniprint",
	Short: "Easily print on a thermal receipt printer",
	Long:  `A flexible thermal printer receipt formatter supporting JSON files and an HTTP API`,
}

var printCmd = &cobra.Command{
	Use:   "print [file]",
	Short: "Print a receipt from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		format, err := printer.ReceiptFormatFromJson(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := printerInterface.QueueReceiptFormat(format); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing receipt: %v\n", err)
			os.Exit(1)
		}

		printerInterface.Execute()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
}

func ExecuteCLI() error {
	return rootCmd.Execute()
}
