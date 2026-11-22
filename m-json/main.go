package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/PaesslerAG/jsonpath"
)

func main() {
	count := flag.Bool("count", false, "Print number of matches")
	raw := flag.Bool("raw", false, "Print matching JSON without indentation")
	pretty := flag.Bool("pretty", false, "Pretty print matching JSON")
	sep := flag.String("sep", "\n", "Separator between results")

	flag.Usage = func() {
		fmt.Println("JSONPath Query Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  m-json [options] input.json jsonpath")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  m-json data.json "$.name"`)
		fmt.Println(`  m-json data.json "$.users[*].name"`)
		fmt.Println(`  m-json data.json "$.users[0]"`)
		fmt.Println(`  m-json data.json "$.users[*]" -count`)
		fmt.Println(`  m-json data.json "$.users[*]" -pretty`)
	}

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	filename := args[0]
	expression := args[1]

	// Read JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(1)
	}

	// Parse JSON
	var document interface{}

	if err := json.Unmarshal(data, &document); err != nil {
		fmt.Fprintln(os.Stderr, "JSON parsing error:", err)
		os.Exit(1)
	}

	// Execute JSONPath
	result, err := jsonpath.Get(expression, document)
	if err != nil {
		fmt.Fprintln(os.Stderr, "JSONPath error:", err)
		os.Exit(1)
	}

	// Normalize result into a list
	var results []interface{}

	switch v := result.(type) {
	case []interface{}:
		results = v
	default:
		results = []interface{}{result}
	}

	if *count {
		fmt.Println(len(results))
		return
	}

	outputs := make([]string, 0, len(results))

	for _, item := range results {
		var output []byte

		if *raw {
			output, err = json.Marshal(item)
		} else if *pretty {
			output, err = json.MarshalIndent(item, "", "  ")
		} else {
			switch v := item.(type) {
			case string:
				fmt.Println(v)
				continue
			case nil:
				output = []byte("null")
			default:
				output, err = json.Marshal(item)
			}
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "Output error:", err)
			os.Exit(1)
		}

		outputs = append(outputs, string(output))
	}

	fmt.Print(strings.Join(outputs, *sep))

	if len(outputs) > 0 {
		fmt.Println()
	}
}
