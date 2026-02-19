package main

import (
	"miniprint/printer"
	"os"
)

var printerInterface printer.Printer

func main() {
	if err := ExecuteCLI(); err != nil {
		os.Exit(1)
	}
}
