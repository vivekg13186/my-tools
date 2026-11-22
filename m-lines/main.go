package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Action struct {
	name   string
	value  string
	filter *regexp.Regexp
}

func main() {
	output := flag.String("o", "", "Output file (default: stdout)")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: lines input-file [actions...] [-o output-file]")
		os.Exit(1)
	}

	input := args[0]

	lines, err := readLines(input)
	if err != nil {
		fatal(err)
	}

	actions, err := parseActions(args[1:])
	if err != nil {
		fatal(err)
	}

	for _, action := range actions {
		lines, err = apply(lines, action)
		if err != nil {
			fatal(err)
		}
	}

	var writer *os.File

	if *output != "" {
		writer, err = os.Create(*output)
		if err != nil {
			fatal(err)
		}
		defer writer.Close()
	} else {
		writer = os.Stdout
	}

	w := bufio.NewWriter(writer)

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	if err := w.Flush(); err != nil {
		fatal(err)
	}
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func parseActions(args []string) ([]Action, error) {
	var actions []Action
	var filter *regexp.Regexp

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-if" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-if requires a regex")
			}

			i++
			var err error
			filter, err = regexp.Compile(args[i])
			if err != nil {
				return nil, fmt.Errorf("invalid regex: %w", err)
			}
			continue
		}

		action := Action{
			name:   arg,
			filter: filter,
		}

		filter = nil

		switch arg {
		case "asc", "dsc", "dedup", "rm-empty", "trim":

		case "prefix", "suffix", "filter":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires text", arg)
			}
			i++
			action.value = args[i]

		default:
			return nil, fmt.Errorf("unknown action: %s", arg)
		}

		actions = append(actions, action)
	}

	if filter != nil {
		return nil, fmt.Errorf("-if must be followed by an action")
	}

	return actions, nil
}

func matches(line string, filter *regexp.Regexp) bool {
	return filter == nil || filter.MatchString(line)
}

func apply(lines []string, a Action) ([]string, error) {
	switch a.name {

	case "trim", "prefix", "suffix":
		for i, line := range lines {
			if !matches(line, a.filter) {
				continue
			}

			switch a.name {
			case "trim":
				lines[i] = strings.TrimSpace(line)
			case "prefix":
				lines[i] = a.value + line
			case "suffix":
				lines[i] = line + a.value
			}
		}

	case "rm-empty":
		result := make([]string, 0, len(lines))

		for _, line := range lines {
			if matches(line, a.filter) && strings.TrimSpace(line) == "" {
				continue
			}
			result = append(result, line)
		}

		lines = result

	case "dedup":
		seen := make(map[string]bool)
		result := make([]string, 0, len(lines))

		for _, line := range lines {
			if matches(line, a.filter) {
				if seen[line] {
					continue
				}
				seen[line] = true
			}
			result = append(result, line)
		}

		lines = result

	case "asc", "dsc":
		var positions []int
		var selected []string

		for i, line := range lines {
			if matches(line, a.filter) {
				positions = append(positions, i)
				selected = append(selected, line)
			}
		}

		sort.Strings(selected)

		if a.name == "dsc" {
			sort.Sort(sort.Reverse(sort.StringSlice(selected)))
		}

		for i, position := range positions {
			lines[position] = selected[i]
		}
	case "filter":
		re, err := regexp.Compile(a.value)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex: %w", err)
		}

		result := make([]string, 0, len(lines))

		for _, line := range lines {
			if matches(line, a.filter) && re.MatchString(line) {
				result = append(result, line)
			}
		}

		lines = result
	default:
		return nil, fmt.Errorf("unsupported action: %s", a.name)
	}

	return lines, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
