package main

func main() {
	printer, ctx := GetPrinter()
	defer ctx.Close()

	ResetPrinter(printer)

	PrintText(
		printer,
		"TOTAL £123.45",
		FontA,
		true,
		AlignCenter,
	)

	LineBreak(printer)

	PrintText(
		printer,
		"Something Funny",
		FontB,
		false,
		AlignLeft,
	)

	LineBreak(printer)

	largeTable, _ := NewTable(
		FontA,
		[]TableColumn{
			{Header: "Item", WidthPct: 60},
			{Header: "Qty", WidthPct: 10},
			{Header: "Price", WidthPct: 30},
		},
	)

	largeTable.AddRow("Apples", "2", "£1.20")
	largeTable.AddRow("Very long product name that wraps nicely", "10", "£9.99")
	largeTable.AddRow("Something else ig", "900", "£13.00")

	largeTable.Print(printer)

	LineBreak(printer)

	smallTable, _ := NewTable(
		FontB,
		[]TableColumn{
			{Header: "Item", WidthPct: 60},
			{Header: "Qty", WidthPct: 10},
			{Header: "Price", WidthPct: 30},
		},
	)

	smallTable.AddRow("Apples", "2", "£1.20")
	smallTable.AddRow("Very long product name that wraps nicely", "10", "£9.99")
	smallTable.AddRow("Something else ig", "900", "£13.00")

	smallTable.Print(printer)

	printer.PrintAndCut()
}
