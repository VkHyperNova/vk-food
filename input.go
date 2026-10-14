package main

import (
	"github.com/peterh/liner"
)

var line = liner.NewLiner()

func PromptWithSuggestion(name string, edit string) (string, error) {
	return line.PromptWithSuggestion(name, edit, -1)
}

func CloseInput() {
	line.Close()
}
