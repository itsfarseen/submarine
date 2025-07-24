package main

import (
	"fmt"
	"log"
	"submarine/codegen"
)

func main() {
	// Define the directory containing the YAML type definitions.
	yamlDir := "codegen/yaml"
	outputDir := "generated"

	// Parse all the YAML files into a structured format.
	allModules, err := codegen.Parse(yamlDir)
	if err != nil {
		log.Fatalf("Error parsing modules from %s: %v", yamlDir, err)
	}

	fmt.Println("Successfully parsed all modules. Starting validation...")

	// Validate the parsed modules to check for broken references or imports.
	codegen.Validate(allModules)

	fmt.Println("Validation complete. Starting code generation...")

	if err := codegen.Generate(allModules, outputDir); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}

	fmt.Println("Code generation complete.")
}
