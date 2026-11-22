package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	url := os.Args[1]
	selector := ""
	outputPath := ""
	useXPath := false

	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-p":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "Error: -p requires a file path")
				os.Exit(1)
			}
			i++
			outputPath = os.Args[i]

		case "--xpath", "-xpath":
			useXPath = true

		case "--css", "-css":
			useXPath = false

		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
				os.Exit(1)
			}
			if selector != "" {
				fmt.Fprintln(os.Stderr, "Error: only one selector is supported")
				os.Exit(1)
			}
			selector = arg
		}
	}

	// Fetch webpage
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Goget/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fatal(fmt.Errorf("HTTP error: %s", resp.Status))
	}

	// Read HTML
	var html bytes.Buffer
	if _, err := html.ReadFrom(resp.Body); err != nil {
		fatal(err)
	}

	var extracted string

	if useXPath {
		extracted, err = extractXPath(html.Bytes(), selector)
	} else {
		extracted, err = extractCSS(html.Bytes(), selector)
	}

	if err != nil {
		fatal(err)
	}

	extracted = normalizeText(extracted)

	if extracted == "" {
		fmt.Println("No text found.")
		return
	}

	// Print extracted text
	fmt.Println(extracted)

	// Optionally save to file
	if outputPath != "" {
		err := os.WriteFile(
			outputPath,
			[]byte(extracted),
			0644,
		)
		if err != nil {
			fatal(err)
		}

		fmt.Fprintf(os.Stderr, "\nSaved to: %s\n", outputPath)
	}
}

func extractCSS(data []byte, selector string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	if selector == "" {
		selector = "body"
	}

	return doc.Find(selector).Text(), nil
}

func extractXPath(data []byte, expression string) (string, error) {
	if expression == "" {
		expression = "//body"
	}

	doc, err := htmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	nodes, err := htmlquery.QueryAll(doc, expression)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	for _, node := range nodes {
		result.WriteString(htmlquery.InnerText(node))
		result.WriteString("\n")
	}

	return result.String(), nil
}

func normalizeText(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(strings.Join(strings.Fields(line), " "))
		if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func usage() {
	fmt.Println(`Usage:
  get URL [selector] [options]

Options:
  --css       Use CSS selector (default)
  --xpath     Use XPath selector
  -p PATH     Save extracted text to file

Examples:
  get https://example.com
  get https://example.com ".content"
  get https://example.com "//article" --xpath
  get https://example.com "article" -p output.txt`)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
