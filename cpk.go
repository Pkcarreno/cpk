package main

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/ui"
)

func run(arguments []string, stdout io.Writer, stderr io.Writer) int {
	rootCommand := cmd.NewRootCommand()
	rootCommand.SetArgs(arguments)
	rootCommand.SetOut(stdout)
	rootCommand.SetErr(stderr)

	executionError := rootCommand.Execute()
	if executionError == nil {
		return 0
	}

	isVerbose := slices.Contains(arguments, "-v") || slices.Contains(arguments, "--verbose")
	if flag := rootCommand.PersistentFlags().Lookup("verbose"); flag != nil && flag.Value.String() == "true" {
		isVerbose = true
	}

	hasNoColorFlag := slices.Contains(arguments, "--no-color")
	if flag := rootCommand.PersistentFlags().Lookup("no-color"); flag != nil && flag.Value.String() == "true" {
		hasNoColorFlag = true
	}

	isTerminal := ui.IsTerminal(stderr) && !ui.IsCI()
	noColor := hasNoColorFlag || ui.HasNoColor() || !isTerminal

	formattedError := ui.FormatError(executionError, isVerbose, noColor)
	fmt.Fprint(stderr, formattedError)

	return 1
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
