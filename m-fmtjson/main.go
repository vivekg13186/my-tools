package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	compact := flag.Bool("compact", false, "Output compact JSON")
	indent := flag.Int("indent", 2, "Number of indentation spaces")
	flag.Usage = func() {
		fmt.Println("JSON Formatter")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  m-fmtjson [options] input.json [output.json]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 || len(args) > 2 {
		flag.Usage()
		os.Exit(1)
	}

	if *indent < 0 || *indent > 16 {
		fmt.Fprintln(os.Stderr, "Error: indent must be between 0 and 16")
		os.Exit(1)
	}

	inputFile := args[0]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(1)
	}

	var output bytes.Buffer

	if *compact {
		err = json.Compact(&output, data)
	} else {
		prefix := ""
		spaces := strings.Repeat(" ", *indent)
		if *indent == 0 {
			spaces = ""
		}
		err = json.Indent(&output, data, prefix, spaces)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid JSON:", err)
		os.Exit(1)
	}

	if !*compact {
		output.WriteByte('\n')
	}

	result := output.Bytes()

	if len(args) == 2 {
		if err := os.WriteFile(args[1], result, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "Error writing file:", err)
			os.Exit(1)
		}
		fmt.Println("JSON formatted:", args[1])
	} else {
		os.Stdout.Write(result)
	}
}
