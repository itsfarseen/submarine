package main

import (
	"fmt"
	"log"
	"path/filepath"
	"submarine/codegen"
	"submarine/codegen/parser"
)

func main() {
	// Define the directory containing the YAML type definitions.
	yamlDir := "codegen/yaml"
	rootModulePath := "submarine/scale/gen"
	outputDir := "scale/gen"

	files := []string{
		"scaleInfo.yaml",
		"v9.yaml",
		"v10.yaml",
		"v11.yaml",
		"v12.yaml",
		"v13.yaml",
		"v14.yaml",
	}

	for i := range files {
		files[i] = filepath.Join(yamlDir, files[i])
	}

	// Parse all the YAML files into a structured format.
	allModules, err := parser.Parse(files)
	if err != nil {
		log.Fatalf("Error parsing modules from %s: %v", yamlDir, err)
	}

	fmt.Println("Successfully parsed all modules. Starting validation...")

	// Validate the parsed modules to check for broken references or imports.
	parser.Validate(allModules)

	fmt.Println("Validation complete. Starting code generation...")

	if err := codegen.Generate(allModules, rootModulePath, outputDir); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}

	fmt.Println("Code generation complete.")
}
