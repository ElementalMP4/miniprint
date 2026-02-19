package main

import (
	"fmt"
	"miniprint/printer"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var replacements []string

var rootCmd = &cobra.Command{
	Use:   "miniprint",
	Short: "Easily print on a thermal receipt printer",
	Long:  `A flexible thermal printer receipt formatter supporting JSON files and an HTTP API`,
}

func applyReplacements(input string, pairs []string) (string, error) {
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid --replace value %q (expected KEY=VALUE)", pair)
		}

		key := parts[0]
		value := parts[1]

		placeholder := "{{" + key + "}}"
		input = strings.ReplaceAll(input, placeholder, value)
	}
	return input, nil
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

		jsonStr, err := applyReplacements(string(data), replacements)
		if err != nil {
			return err
		}

		format, err := printer.ReceiptFormatFromJson(jsonStr)
		if err != nil {
			return fmt.Errorf("error loading JSON file: %w", err)
		}

		if err := printerInterface.QueueReceiptFormat(format); err != nil {
			return fmt.Errorf("error printing receipt: %w", err)
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

		jsonStr, err := applyReplacements(string(data), replacements)
		if err != nil {
			return err
		}

		format, err := printer.ReceiptFormatFromJson(jsonStr)
		if err != nil {
			return fmt.Errorf("error loading JSON file: %w", err)
		}

		if err := printer.RenderReceiptPreviewCanvas(format, "preview.png"); err != nil {
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
	printCmd.Flags().StringArrayVar(
		&replacements,
		"replace",
		nil,
		"Replace template variables (KEY=VALUE)",
	)

	previewCmd.Flags().StringArrayVar(
		&replacements,
		"replace",
		nil,
		"Replace template variables (KEY=VALUE)",
	)

	rootCmd.AddCommand(printCmd)
	rootCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(cutCmd)
	rootCmd.AddCommand(serveCmd)
}

func ExecuteCLI() error {
	return rootCmd.Execute()
}
