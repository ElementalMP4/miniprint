package main

import (
	"miniprint/printer"
	"os"
)

var printerInterface printer.Printer

func main() {
	printerInterface.Initialise()

	if err := ExecuteCLI(); err != nil {
		os.Exit(1)
	}
}
