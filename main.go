package main

import (
	"io"
	"log"
	"os"

	"github.com/yuin/goldmark"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
	}
	err = goldmark.Convert(input, os.Stdout)
	if err != nil {
		log.Fatalf("failed to convert: %v", err)
	}
}
