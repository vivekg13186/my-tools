package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/antchfx/xmlquery"
)

func main() {
	count := flag.Bool("count", false, "Print number of matches")
	raw := flag.Bool("raw", false, "Print matching XML instead of text")
	attr := flag.String("attr", "", "Print a specific attribute")
	separator := flag.String("sep", "\n", "Separator between results")

	flag.Usage = func() {
		fmt.Println("XML XPath Query Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  m-xml [options] input.xml xpath")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println(`  m-xml data.xml "//book"`)
		fmt.Println(`  m-xml data.xml "//book/title"`)
		fmt.Println(`  m-xml data.xml "//book[@id='1']"`)
		fmt.Println(`  m-xml data.xml "//book" -raw`)
		fmt.Println(`  m-xml data.xml "//book" -attr id`)
		fmt.Println(`  m-xml data.xml "//book" -count`)
	}

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	filename := args[0]
	expression := args[1]

	// Read XML file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse XML
	doc, err := xmlquery.Parse(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "XML parsing error:", err)
		os.Exit(1)
	}

	// Execute XPath
	nodes, err := xmlquery.QueryAll(doc, expression)
	if err != nil {
		fmt.Fprintln(os.Stderr, "XPath error:", err)
		os.Exit(1)
	}

	if *count {
		fmt.Println(len(nodes))
		return
	}

	if len(nodes) == 0 {
		return
	}

	results := make([]string, 0, len(nodes))

	for _, node := range nodes {
		var result string

		switch {
		case *attr != "":
			// Extract requested attribute
			for _, a := range node.Attr {
				if a.Name.Local == *attr {
					result = a.Value
					break
				}
			}

		case *raw:
			// Return complete XML node
			result = node.OutputXML(true)

		default:
			// Return text content
			result = node.InnerText()

			// Handle XPath attribute/text selections
			if result == "" {
				result = node.Data
			}
		}

		results = append(results, result)
	}

	fmt.Print(strings.Join(results, *separator))

	if len(results) > 0 && !strings.HasSuffix(results[len(results)-1], "\n") {
		fmt.Println()
	}
}
