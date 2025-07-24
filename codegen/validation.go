package codegen

import "fmt"

// a set of known primitive types that are valid references.
var primitives = map[string]struct{}{
	"text":             {},
	"bytes":            {},
	"u8":               {},
	"u32":              {},
	"u64":              {},
	"bool":             {},
	"type":             {}, // 'type' is used as a generic type placeholder in some definitions
	"compact":          {},
	"Si1LookupTypeId":  {},
	"Si1Type":          {},
	"PortableRegistry": {},
}

// Validate checks the integrity of all parsed modules. It verifies that
// all internal type references (Ref) and cross-module imports (Import)
// are valid and point to existing definitions.
func Validate(allModules *AllModules) {
	validateRefs(allModules)
	validateImports(allModules)
}

func validateImports(allModules *AllModules) {
	for moduleName, module := range allModules.Modules {
		for typeName, typeDef := range module.Types {
			if typeDef.Kind == KindImport {
				targetModuleName := typeDef.Import.Module
				targetTypeName := typeDef.Import.Item

				targetModule, ok := allModules.Modules[targetModuleName]
				if !ok {
					fmt.Printf("Validation Error in module '%s': Type '%s' imports from non-existent module '%s'\n",
						moduleName, typeName, targetModuleName)
					continue
				}

				if _, ok := targetModule.Types[targetTypeName]; !ok {
					fmt.Printf("Validation Error in module '%s': Type '%s' imports non-existent type '%s' from module '%s'\n",
						moduleName, typeName, targetTypeName, targetModuleName)
				}
			}
		}
	}
}

func validateRefs(allModules *AllModules) {
	for moduleName, module := range allModules.Modules {
		for typeName, typeDef := range module.Types {
			walkAndValidateRefs(allModules, moduleName, typeName, typeDef, make(map[string]bool))
		}
	}
}

// walkAndValidateRefs recursively traverses a type definition to find all
// Ref kinds and validates them.
func walkAndValidateRefs(allModules *AllModules, moduleName, currentTypeName string, typeDef *Type, visited map[string]bool) {
	// Prevent infinite recursion on circular dependencies.
	if visited[currentTypeName] {
		return
	}
	visited[currentTypeName] = true

	switch typeDef.Kind {
	case KindRef:
		refName := typeDef.Ref.Name
		// Check if the reference is a primitive.
		if _, isPrimitive := primitives[refName]; isPrimitive {
			return
		}
		// Check if the reference exists in the current module's types.
		if _, ok := allModules.Modules[moduleName].Types[refName]; !ok {
			fmt.Printf("Validation Error in module '%s': Type '%s' has an unresolved reference to '%s'\n",
				moduleName, currentTypeName, refName)
		}
	case KindStruct:
		for _, field := range typeDef.Struct.Fields {
			walkAndValidateRefs(allModules, moduleName, field.Name, field.Type, visited)
		}
	case KindEnumComplex:
		for _, variant := range typeDef.EnumComplex.Variants {
			walkAndValidateRefs(allModules, moduleName, variant.Name, variant.Type, visited)
		}
	case KindVec, KindOption:
		var innerType *Type
		if typeDef.Kind == KindVec {
			innerType = typeDef.Vec.Type
		} else {
			innerType = typeDef.Option.Type
		}
		walkAndValidateRefs(allModules, moduleName, currentTypeName, innerType, visited)
	}
}
