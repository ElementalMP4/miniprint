package main

import (
	"fmt"
	"miniprint/printer"
	"os"
)

var printerInterface printer.Printer

func main() {
	printerInterface.Initialise()

	if err := ExecuteCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
