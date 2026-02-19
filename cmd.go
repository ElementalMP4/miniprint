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
		printerInterface.Initialise()
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

		if format.Settings.NoCut {
			printerInterface.ExecuteWithoutCut()
		} else {
			printerInterface.Execute()
		}
		return nil
	},
}

var previewCmd = &cobra.Command{
	Use:   "preview [file]",
	Short: "Preview a receipt from a JSON file",
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

		err = printer.RenderReceiptPreviewCanvas(format, "preview.png")
		if err != nil {
			return err
		}

		fmt.Println("Generated preview")
		return nil
	},
}

var cutCmd = &cobra.Command{
	Use:   "cut",
	Short: "Cut the receipt paper without adding any new content",
	RunE: func(cmd *cobra.Command, args []string) error {
		printerInterface.Initialise()
		printerInterface.Cut()
		return nil
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start an HTTP server that allows communication with the printer",
	RunE: func(cmd *cobra.Command, args []string) error {
		printerInterface.Initialise()
		return serve()
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
	rootCmd.AddCommand(cutCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(previewCmd)
}

func ExecuteCLI() error {
	return rootCmd.Execute()
}
