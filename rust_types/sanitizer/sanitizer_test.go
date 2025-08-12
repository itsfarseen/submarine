package sanitizer_test

import (
	. "submarine/rust_types"
	"submarine/rust_types/sanitizer"
	"testing"
)

func TestNormalizeSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newlines and multiple spaces",
			input:    "Foo\nBar  Baz",
			expected: "Foo Bar Baz",
		},
		{
			name:     "tabs and spaces",
			input:    "Hello\tWorld   Test",
			expected: "Hello World Test",
		},
		{
			name:     "mixed whitespace",
			input:    "A\n\t B\r\n  C",
			expected: "A B C",
		},
		{
			name:     "single space unchanged",
			input:    "Single Space",
			expected: "Single Space",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.NormalizeSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSpaces(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveAsTrait(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple as trait removal",
			input:    "<Foo as Trait>::Bar",
			expected: "Foo::Bar",
		},
		{
			name:     "multiple as trait patterns",
			input:    "<A as TraitA>::<B as TraitB>::method",
			expected: "A::B::method",
		},
		{
			name:     "no as trait pattern",
			input:    "regular::path::Type",
			expected: "regular::path::Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.RemoveAsTrait(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveAsTrait(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseAndSanitizeRustType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "box type removal",
			input:    "Box<String>",
			expected: "text",
		},
		{
			name:     "string to text",
			input:    "String",
			expected: "text",
		},
		{
			name:     "option with generics",
			input:    "Option<i32>",
			expected: "Option<i32>",
		},
		{
			name:     "compact type",
			input:    "Compact<u64>",
			expected: "compact",
		},
		{
			name:     "nested compact type",
			input:    "Vec<Compact<u64>>",
			expected: "Vec<compact>",
		},
		{
			name:     "nested option box vec",
			input:    "Option<Box<Vec<u64>>>",
			expected: "Option<Vec<u64>>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rust_type := sanitizer.ParseRustType(tt.input)
			result := sanitizer.SanitizeRustType(rust_type)
			if result.String() != tt.expected {
				t.Errorf("SanitizeRustType(%q).String() = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeRustType(t *testing.T) {
	tests := []struct {
		name     string
		input    RustType
		expected string
	}{
		{
			name:     "box unwrapping",
			input:    Base([]string{"Box"}, []RustType{Base([]string{"i32"}, nil)}),
			expected: "i32",
		},
		{
			name:     "vec with generics",
			input:    Base([]string{"Vec"}, []RustType{Base([]string{"String"}, nil)}),
			expected: "Vec<text>",
		},
		{
			name:     "tuple sanitization",
			input:    Tuple([]RustType{Base([]string{"i32"}, nil), Base([]string{"String"}, nil)}),
			expected: "(i32, text)",
		},
		{
			name:     "array sanitization",
			input:    Array(Base([]string{"u8"}, nil), 32),
			expected: "[u8; 32]",
		},
		{
			name:     "path simplification",
			input:    Base([]string{"std", "collections", "HashMap"}, []RustType{Base([]string{"String"}, nil), Base([]string{"i32"}, nil)}),
			expected: "std::collections::HashMap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeRustType(tt.input)
			if result.String() != tt.expected {
				t.Errorf("SanitizeRustType(%v).String() = %q, want %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "normalize spaces and remove as trait",
			input:  "<Foo  as\nTrait>::Bar",
			output: "Foo::Bar",
		},
		{
			name:   "complex whitespace",
			input:  "Vec<\t\tT>\t\n",
			output: "Vec<T>",
		},
		{
			name:   "simple type",
			input:  "SimpleType",
			output: "SimpleType",
		},
		{
			name:   "mixed patterns",
			input:  "<A as  B<T>>::C",
			output: "A::C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since Sanitize doesn't return anything, we just test that it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Sanitize(%q) panicked: %v", tt.input, r)
				}
			}()
			output := sanitizer.ParseAndSanitize(tt.input)
			if output.String() != tt.output {
				t.Errorf("ParseAndSanitize(%v).String() = %q, want %q", tt.input, output, tt.output)
			}
		})
	}
}

var GOLDEN_TESTS []struct{ input, output string } = []struct {
	input  string
	output string
}{
	{"<T::AuthorityId as RuntimeAppPublic>::Signature", "AuthorityId::Signature"},
	{"<T::Lookup as StaticLookup>::Source", "Lookup::Source"},
	{"AccountId", "AccountId"},
	{"AccountIndex", "AccountIndex"},
	{"AccountValidity", "AccountValidity"},
	{"AccountVote<BalanceOf<T>>", "AccountVote"}, // <- [x] democracy
	{"Approvals", "Approvals"},
	{"AuctionIndex", "AuctionIndex"},
	{"AuthorityId", "AuthorityId"},
	{"AuthorityList", "AuthorityList"},
	{"Balance", "Balance"},
	{"BalanceOf<T, I>", "Balance"},
	{"BalanceOf<T>", "Balance"},
	{"BlockNumber", "BlockNumber"},
	{"BountyIndex", "BountyIndex"},
	{"Box<<T as Config<I>>::Proposal>", "Proposal"},
	{"Box<<T as Config>::Call>", "Call"},
	{"Box<<T as Trait<I>>::Proposal>", "Proposal"},
	{"Box<<T as Trait>::Call>", "Call"},
	{"Box<EquivocationProof<T::Hash, T::BlockNumber>>", "EquivocationProof"}, // <- [x] grandpa - per-module alias
	{"Box<EquivocationProof<T::Header>>", "EquivocationProof"},               // <- [x] babe - per-module alias
	{"Box<IdentityInfo<T::MaxAdditionalFields>>", "IdentityInfo"},
	{"Box<RawSolution<CompactOf<T>>>", "RawSolution"}, // <-
	{"Box<RawSolution<SolutionOf<T>>>", "RawSolution"},
	{"CallHash", "CallHash"},
	{"CallHashOf<T>", "CallHash"}, // <- [x]
	{"CollatorId", "CollatorId"},
	{"Compact<AuctionIndex>", "compact"},
	{"Compact<BalanceOf<T, I>>", "compact"},
	{"Compact<BalanceOf<T>>", "compact"},
	{"Compact<BountyIndex>", "compact"},
	{"Compact<EraIndex>", "compact"},
	{"Compact<LeasePeriodOf<T>>", "compact"},
	{"Compact<MemberCount>", "compact"},
	{"Compact<ParaId>", "compact"},
	{"Compact<PropIndex>", "compact"},
	{"Compact<ProposalIndex>", "compact"},
	{"Compact<ReferendumIndex>", "compact"},
	{"Compact<RegistrarIndex>", "compact"},
	{"Compact<SubId>", "compact"},
	{"Compact<T::Balance>", "compact"},
	{"Compact<T::BlockNumber>", "compact"},
	{"Compact<T::Moment>", "compact"},
	{"Compact<Weight>", "compact"},
	{"Compact<u32>", "compact"},
	{"CompactAssignments", "CompactAssignments"},
	{"Conviction", "Conviction"},
	{"Data", "Data"},
	{"DefunctVoter<<T::Lookup as StaticLookup>::Source>", "DefunctVoter"}, // <-
	{"DispatchError", "DispatchError"},
	{"DispatchInfo", "DispatchInfo"},
	{"DispatchResult", "DispatchResult"},
	{`DoubleVoteReport<<T::KeyOwnerProofSystem as
                 KeyOwnerProofSystem<(KeyTypeId, ValidatorId)>>::Proof>`, "DoubleVoteReport"}, // <- [x] parachains
	{"EcdsaSignature", "EcdsaSignature"},
	{"ElectionCompute", "ElectionCompute"},
	{"ElectionScore", "ElectionScore"},
	{"ElectionSize", "ElectionSize"},
	{"EquivocationProof<T::Hash, T::BlockNumber>", "EquivocationProof"}, // <-
	{"EquivocationProof<T::Header>", "EquivocationProof"},
	{"EraIndex", "EraIndex"},
	{"EthereumAddress", "EthereumAddress"},
	{"Hash", "Hash"},
	{"HeadData", "HeadData"},
	{"Heartbeat<T::BlockNumber>", "Heartbeat"},
	{"IdentityFields", "IdentityFields"},
	{"IdentityInfo", "IdentityInfo"},
	{"Judgement<BalanceOf<T>>", "Judgement"},
	{"Key", "Key"},
	{"Kind", "Kind"},
	{"LeasePeriod", "LeasePeriod"},
	{"MemberCount", "MemberCount"},
	{"MoreAttestations", "MoreAttestations"},
	{"NewBidder<AccountId>", "NewBidder"},
	{"NextConfigDescriptor", "NextConfigDescriptor"},
	{"OpaqueCall", "OpaqueCall"},
	{"OpaqueTimeSlot", "OpaqueTimeSlot"},
	{"Option<(BalanceOf<T>, BalanceOf<T>, T::BlockNumber)>", "Option<(Balance, Balance, BlockNumber)>"},
	{"Option<ChangesTrieConfiguration>", "Option<ChangesTrieConfiguration>"},
	{"Option<ElectionCompute>", "Option<ElectionCompute>"},
	{"Option<ElectionScore>", "Option<ElectionScore>"},
	{"Option<Percent>", "Option<Percent>"},
	{"Option<ReferendumIndex>", "Option<ReferendumIndex>"},
	{"Option<StatementKind>", "Option<StatementKind>"},
	{"Option<T::AccountId>", "Option<AccountId>"},
	{"Option<T::ProxyType>", "Option<ProxyType>"},              // <-
	{"Option<Timepoint<T::BlockNumber>>", "Option<Timepoint>"}, // <-
	{"Option<schedule::Period<T::BlockNumber>>", "Option<schedule::Period>"},
	{"Option<u32>", "Option<u32>"},
	{"ParaId", "ParaId"},
	{"ParaInfo", "ParaInfo"},
	{"Perbill", "Perbill"},
	{"Percent", "Percent"},
	{"Permill", "Permill"},
	{"PhragmenScore", "PhragmenScore"},
	{"PropIndex", "PropIndex"},
	{"ProposalIndex", "ProposalIndex"},
	{"ProxyType", "ProxyType"},
	{"RawSolution<CompactOf<T>>", "RawSolution"},     // <- [x] staking
	{"ReadySolution<T::AccountId>", "ReadySolution"}, // <- [x] staking
	{"ReferendumIndex", "ReferendumIndex"},
	{"RegistrarIndex", "RegistrarIndex"},
	{"Remark", "Remark"},
	{"Renouncing", "Renouncing"},
	{"RewardDestination", "RewardDestination"},
	{"RewardDestination<T::AccountId>", "RewardDestination"}, // <- [x] staking
	{"SessionIndex", "SessionIndex"},
	{"SlotRange", "SlotRange"},
	{"SolutionOrSnapshotSize", "SolutionOrSnapshotSize"},
	{"Status", "Status"},
	{"Supports<T::AccountId>", "Supports"}, // <---- [x] staking
	{"T::AccountId", "AccountId"},
	{"T::AccountIndex", "AccountIndex"},
	{"T::BlockNumber", "BlockNumber"},
	{"T::Hash", "Hash"}, // <-
	{"T::KeyOwnerProof", "KeyOwnerProof"},
	{"T::Keys", "Keys"},
	{"T::ProxyType", "ProxyType"},
	{"TaskAddress<BlockNumber>", "TaskAddress"},
	{"Timepoint<BlockNumber>", "Timepoint"},
	{"Timepoint<T::BlockNumber>", "Timepoint"},
	{"ValidationCode", "ValidationCode"},
	{"ValidatorPrefs", "ValidatorPrefs"},
	{"Vec<(AccountId, Balance)>", "Vec<(AccountId, Balance)>"},
	{"Vec<(T::AccountId, Data)>", "Vec<(AccountId, Data)>"},
	{"Vec<(T::AccountId, u32)>", "Vec<(AccountId, u32)>"},
	{"Vec<<T as Config>::Call>", "Vec<Call>"},
	{"Vec<<T as Trait>::Call>", "Vec<Call>"},
	{"Vec<<T::Lookup as StaticLookup>::Source>", "Vec<Lookup::Source>"},
	{"Vec<AccountId>", "Vec<AccountId>"},
	{"Vec<AttestedCandidate>", "Vec<AttestedCandidate>"},
	{"Vec<IdentificationTuple>", "Vec<IdentificationTuple>"},
	{"Vec<Key>", "Vec<Key>"},
	{"Vec<KeyValue>", "Vec<KeyValue>"},
	{"Vec<T::AccountId>", "Vec<AccountId>"},
	{"Vec<T::Header>", "Vec<Header>"},
	{"Vec<ValidatorIndex>", "Vec<ValidatorIndex>"},
	{"Vec<u32>", "Vec<u32>"},
	{"VestingInfo<BalanceOf<T>, T::BlockNumber>", "VestingInfo"},
	{"VoteThreshold", "VoteThreshold"},
	{"Weight", "Weight"},
	{"[u8; 32]", "[u8; 32]"},
	{"bool", "bool"},
	{"schedule::Priority", "schedule::Priority"},
	{"sp_std::marker::PhantomData<(AccountId, Event)>", "empty"}, // <--- [x]
	{"u16", "u16"},
	{"u32", "u32"},
	{"u64", "u64"},
}

func TestParseAndSanitizeGolden(t *testing.T) {
	for _, tt := range GOLDEN_TESTS {
		t.Run(tt.input, func(t *testing.T) {
			// Since Sanitize doesn't return anything, we just test that it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Sanitize(%q) panicked: %v", tt.input, r)
				}
			}()
			output := sanitizer.ParseAndSanitize(tt.input)
			if output.String() != tt.output {
				t.Errorf("ParseAndSanitize(%v).String() = %q, want %q", tt.input, output, tt.output)
			}
		})
	}
}
