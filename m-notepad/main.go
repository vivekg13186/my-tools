package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: text <filename> <content>")
		os.Exit(1)
	}

	filename := os.Args[1]
	content := strings.Join(os.Args[2:], " ")

	// Write content to file
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("File created:", filename)
}
