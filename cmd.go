package main

import (
	"fmt"
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
	Run: func(cmd *cobra.Command, args []string) {
		printer, ctx := GetPrinter()
		defer ctx.Close()
		ResetPrinter(printer)

		format, err := LoadReceiptFormat(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := PrintReceipt(printer, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing receipt: %v\n", err)
			os.Exit(1)
		}

		printer.PrintAndCut()
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
}

func ExecuteCLI() error {
	return rootCmd.Execute()
}
