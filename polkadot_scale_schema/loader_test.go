package polkadot_scale_schema_test

import (
	"submarine/polkadot_scale_schema"
	s "submarine/scale"
	"testing"
)

func TestPolkadotSchemaLoadAndLookup(t *testing.T) {
	// Test schema data that covers all the different type patterns
	testSchema := map[string]map[string]any{
		"runtime": {
			// Basic types
			"AccountId":   "Vec<u8>",
			"Balance":     "u128",
			"BlockNumber": "u32",
			"Hash":        "[u8; 32]",

			// Struct types
			"Transfer": map[string]any{
				"to":     "AccountId",
				"amount": "Balance",
			},

			// Simple enum
			"Verdict": map[string]any{
				"_enum": []any{"Innocent", "Guilty"},
			},

			// Complex enum with variants
			"MultiAddress": map[string]any{
				"_enum": map[string]any{
					"Id":      "AccountId",
					"Index":   "u32",
					"Raw":     "Vec<u8>",
					"Address": nil,
				},
			},

			// Bitflags
			"Permissions": map[string]any{
				"_set": map[string]any{
					"Read":    1,
					"Write":   2,
					"Execute": 4,
					"Admin":   8,
				},
				"_bitLength": 8,
			},

			// Nested generics
			"Result": "Result<Balance, u32>",

			// Option types
			"MaybeBalance": "Option<Balance>",

			// Empty tuple
			"Unit": map[string]any{},

			// Non-empty tuple would be represented as array in JSON
			"Pair": "(u32, u64)",
		},

		"pallet_balances": {
			"AccountId": "Vec<u8>", // Same name, different module
			"Transfer": map[string]any{
				"from":   "AccountId",
				"to":     "AccountId",
				"amount": "u128",
			},
		},
	}

	tests := []struct {
		name         string
		moduleName   string
		typeName     string
		expectedType s.Type
		shouldError  bool
	}{
		{
			name:       "basic u128 type",
			moduleName: "runtime",
			typeName:   "Balance",
			expectedType: s.Type{
				Kind: s.KindRef,
				Ref:  stringPtr("u128"),
			},
		},
		{
			name:       "Vec<u8> type",
			moduleName: "runtime",
			typeName:   "AccountId",
			expectedType: s.Type{
				Kind: s.KindVec,
				Vec: &s.Vec{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("u8"),
					},
				},
			},
		},
		{
			name:       "array type [u8; 32]",
			moduleName: "runtime",
			typeName:   "Hash",
			expectedType: s.Type{
				Kind: s.KindArray,
				Array: &s.Array{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("u8"),
					},
					Len: 32,
				},
			},
		},
		{
			name:       "struct type",
			moduleName: "runtime",
			typeName:   "Transfer",
			expectedType: s.Type{
				Kind: s.KindStruct,
				Struct: &s.Struct{
					Fields: []s.NamedMember{
						{Name: "to", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("AccountId")}},
						{Name: "amount", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("Balance")}},
					},
				},
			},
		},
		{
			name:       "simple enum",
			moduleName: "runtime",
			typeName:   "Verdict",
			expectedType: s.Type{
				Kind: s.KindEnumSimple,
				EnumSimple: &s.EnumSimple{
					Variants: []string{"Innocent", "Guilty"},
				},
			},
		},
		{
			name:       "complex enum",
			moduleName: "runtime",
			typeName:   "MultiAddress",
			expectedType: s.Type{
				Kind: s.KindEnumComplex,
				EnumComplex: &s.EnumComplex{
					Variants: []s.NamedMember{
						{Name: "Id", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("AccountId")}},
						{Name: "Index", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("u32")}},
						{Name: "Raw", Type: &s.Type{Kind: s.KindVec, Vec: &s.Vec{Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("u8")}}}},
						{Name: "Address", Type: nil},
					},
				},
			},
		},
		{
			name:       "bitflags type",
			moduleName: "runtime",
			typeName:   "Permissions",
			expectedType: s.Type{
				Kind: s.KindBitFlags,
				BitFlags: &s.BitFlags{
					BitLength: 8,
					Flags: []s.BitFlag{
						{Name: "Read", Value: 1},
						{Name: "Write", Value: 2},
						{Name: "Execute", Value: 4},
						{Name: "Admin", Value: 8},
					},
				},
			},
		},
		{
			name:       "Option type",
			moduleName: "runtime",
			typeName:   "MaybeBalance",
			expectedType: s.Type{
				Kind: s.KindOption,
				Option: &s.Option{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("Balance"),
					},
				},
			},
		},
		{
			name:       "empty struct/unit type",
			moduleName: "runtime",
			typeName:   "Unit",
			expectedType: s.Type{
				Kind: s.KindStruct,
				Struct: &s.Struct{
					Fields: []s.NamedMember{},
				},
			},
		},
		{
			name:       "same type name different module",
			moduleName: "pallet_balances",
			typeName:   "AccountId",
			expectedType: s.Type{
				Kind: s.KindVec,
				Vec: &s.Vec{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("u8"),
					},
				},
			},
		},
		{
			name:       "different struct same name different module",
			moduleName: "pallet_balances",
			typeName:   "Transfer",
			expectedType: s.Type{
				Kind: s.KindStruct,
				Struct: &s.Struct{
					Fields: []s.NamedMember{
						{Name: "from", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("AccountId")}},
						{Name: "to", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("AccountId")}},
						{Name: "amount", Type: &s.Type{Kind: s.KindRef, Ref: stringPtr("u128")}},
					},
				},
			},
		},
		{
			name:        "type not found",
			moduleName:  "runtime",
			typeName:    "NonExistent",
			shouldError: true,
		},
	}

	// Load the test schema into registry
	registry := polkadot_scale_schema.NewRegistry()
	err := polkadot_scale_schema.LoadFromSchema(registry, testSchema)
	if err != nil {
		t.Fatalf("Failed to load test schema: %v", err)
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lazyType, err := registry.Lookup(tt.moduleName, tt.typeName)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, but got no error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			actualType, err := lazyType.ToScaleType()
			if err != nil {
				t.Fatalf("Failed to convert to scale type: %v", err)
			}

			if !typesEqual(*actualType, tt.expectedType) {
				t.Errorf("Type mismatch.\nExpected: %+v\nActual: %+v", tt.expectedType, *actualType)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Deep comparison function for s.Type
func typesEqual(a, b s.Type) bool {
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case s.KindRef:
		return *a.Ref == *b.Ref

	case s.KindVec:
		return typesEqual(*a.Vec.Type, *b.Vec.Type)

	case s.KindArray:
		return a.Array.Len == b.Array.Len && typesEqual(*a.Array.Type, *b.Array.Type)

	case s.KindOption:
		return typesEqual(*a.Option.Type, *b.Option.Type)

	case s.KindTuple:
		if len(a.Tuple.Fields) != len(b.Tuple.Fields) {
			return false
		}
		for i := range a.Tuple.Fields {
			if !typesEqual(a.Tuple.Fields[i], b.Tuple.Fields[i]) {
				return false
			}
		}
		return true

	case s.KindStruct:
		return namedMembersEqual(a.Struct.Fields, b.Struct.Fields)

	case s.KindEnumSimple:
		return stringSlicesEqual(a.EnumSimple.Variants, b.EnumSimple.Variants)

	case s.KindEnumComplex:
		return namedMembersEqual(a.EnumComplex.Variants, b.EnumComplex.Variants)

	case s.KindBitFlags:
		if a.BitFlags.BitLength != b.BitFlags.BitLength {
			return false
		}
		return bitFlagsEqual(a.BitFlags.Flags, b.BitFlags.Flags)

	default:
		return false
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func namedMembersEqual(a, b []s.NamedMember) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison since order might not matter
	aMap := make(map[string]*s.Type)
	bMap := make(map[string]*s.Type)

	for _, member := range a {
		aMap[member.Name] = member.Type
	}
	for _, member := range b {
		bMap[member.Name] = member.Type
	}

	if len(aMap) != len(bMap) {
		return false
	}

	for name, aType := range aMap {
		bType, exists := bMap[name]
		if !exists {
			return false
		}

		// Handle nil types
		if aType == nil && bType == nil {
			continue
		}
		if aType == nil || bType == nil {
			return false
		}

		if !typesEqual(*aType, *bType) {
			return false
		}
	}

	return true
}

func bitFlagsEqual(a, b []s.BitFlag) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison since order might not matter
	aMap := make(map[string]uint64)
	bMap := make(map[string]uint64)

	for _, flag := range a {
		aMap[flag.Name] = flag.Value
	}
	for _, flag := range b {
		bMap[flag.Name] = flag.Value
	}

	if len(aMap) != len(bMap) {
		return false
	}

	for name, aValue := range aMap {
		bValue, exists := bMap[name]
		if !exists || aValue != bValue {
			return false
		}
	}

	return true
}

func TestRegistryEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		schema      map[string]map[string]any
		moduleName  string
		typeName    string
		shouldError bool
	}{
		{
			name:        "empty schema",
			schema:      map[string]map[string]any{},
			moduleName:  "runtime",
			typeName:    "Test",
			shouldError: true,
		},
		{
			name: "invalid enum variant type",
			schema: map[string]map[string]any{
				"test": {
					"BadEnum": map[string]any{
						"_enum": []any{"Valid", 123}, // Invalid variant type
					},
				},
			},
			moduleName:  "test",
			typeName:    "BadEnum",
			shouldError: true,
		},
		{
			name: "invalid set flag value",
			schema: map[string]map[string]any{
				"test": {
					"BadSet": map[string]any{
						"_set": map[string]any{
							"Flag": "invalid", // Should be number
						},
					},
				},
			},
			moduleName:  "test",
			typeName:    "BadSet",
			shouldError: true,
		},
		{
			name: "unsupported type definition",
			schema: map[string]map[string]any{
				"test": {
					"BadType": []any{1, 2, 3}, // Unsupported array format
				},
			},
			moduleName:  "test",
			typeName:    "BadType",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := polkadot_scale_schema.NewRegistry()
			err := polkadot_scale_schema.LoadFromSchema(registry, tt.schema)

			if err != nil && tt.shouldError {
				return
			}

			if err != nil {
				t.Fatalf("Unexpected schema load error: %v", err)
			}

			_, err = registry.Lookup(tt.moduleName, tt.typeName)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
			} else if err != nil {
				t.Errorf("Unexpected lookup error: %v", err)
			}
		})
	}
}

func TestRegistryMethods(t *testing.T) {
	schema := map[string]map[string]any{
		"runtime": {
			"Balance": "u128",
			"Hash":    "[u8; 32]",
		},
		"balances": {
			"Balance": "u64", // Same name, different module
			"Account": "Vec<u8>",
		},
	}

	registry := polkadot_scale_schema.NewRegistry()
	err := polkadot_scale_schema.LoadFromSchema(registry, schema)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Test GetAllTypes
	t.Run("GetAllTypes", func(t *testing.T) {
		entries, exists := registry.GetAllTypes("Balance")
		if !exists {
			t.Error("Expected Balance type to exist")
			return
		}
		if len(entries) != 2 {
			t.Errorf("Expected 2 entries for Balance, got %d", len(entries))
		}

		// Check both modules are represented
		moduleFound := make(map[string]bool)
		for _, entry := range entries {
			moduleFound[entry.Module] = true
		}
		if !moduleFound["runtime"] || !moduleFound["balances"] {
			t.Error("Expected both runtime and balances modules for Balance type")
		}
	})

	// Test GetModuleTypes
	t.Run("GetModuleTypes", func(t *testing.T) {
		runtimeTypes := registry.GetModuleTypes("runtime")
		if len(runtimeTypes) != 2 {
			t.Errorf("Expected 2 types in runtime module, got %d", len(runtimeTypes))
		}
		if runtimeTypes["Balance"] == nil || runtimeTypes["Hash"] == nil {
			t.Error("Expected Balance and Hash types in runtime module")
		}

		balancesTypes := registry.GetModuleTypes("balances")
		if len(balancesTypes) != 2 {
			t.Errorf("Expected 2 types in balances module, got %d", len(balancesTypes))
		}
		if balancesTypes["Balance"] == nil || balancesTypes["Account"] == nil {
			t.Error("Expected Balance and Account types in balances module")
		}
	})
}
