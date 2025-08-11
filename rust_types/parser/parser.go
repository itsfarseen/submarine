package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type RustTypeKind int

const (
	KindBase RustTypeKind = iota
	KindTuple
	KindArray
)

type RustType struct {
	Kind  RustTypeKind
	Base  *RustTypeBase
	Tuple *[]RustType
	Array *RustTypeArray
}

type RustTypeBase struct {
	Path     []string
	Generics []RustType // can be empty if no generics
}

// types like [u8; 32] or [foo::Foo<Bar>; 12]
type RustTypeArray struct {
	Base RustType
	Len  int
}

// Constructor functions
func Base(path []string, generics []RustType) RustType {
	return RustType{
		Kind: KindBase,
		Base: &RustTypeBase{
			Path:     path,
			Generics: generics,
		},
	}
}

func Tuple(types []RustType) RustType {
	return RustType{
		Kind:  KindTuple,
		Tuple: &types,
	}
}

func Array(base RustType, len int) RustType {
	return RustType{
		Kind: KindArray,
		Array: &RustTypeArray{
			Base: base,
			Len:  len,
		},
	}
}

func (rt RustType) String() string {
	switch rt.Kind {
	case KindBase:
		if rt.Base == nil {
			return ""
		}
		path := strings.Join(rt.Base.Path, "::")
		if len(rt.Base.Generics) == 0 {
			return path
		}
		var params []string
		for _, param := range rt.Base.Generics {
			params = append(params, param.String())
		}
		return fmt.Sprintf("%s<%s>", path, strings.Join(params, ", "))
	case KindTuple:
		if rt.Tuple == nil {
			return "()"
		}
		var params []string
		for _, param := range *rt.Tuple {
			params = append(params, param.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(params, ", "))
	case KindArray:
		if rt.Array == nil {
			return "[]"
		}
		return fmt.Sprintf("[%s; %d]", rt.Array.Base.String(), rt.Array.Len)
	default:
		return ""
	}
}

type RustTypesParser struct {
	s   string
	pos int
}

func NewRustTypesParser(s string) *RustTypesParser {
	return &RustTypesParser{s: s, pos: 0}
}

func (p *RustTypesParser) Advance() rune {
	if p.pos >= len(p.s) {
		return 0
	}
	r := rune(p.s[p.pos])
	p.pos++
	return r
}

func (p *RustTypesParser) Peek() rune {
	if p.pos >= len(p.s) {
		return 0
	}
	return rune(p.s[p.pos])
}

func (p *RustTypesParser) Matches(prefix string) bool {
	if p.pos+len(prefix) > len(p.s) {
		return false
	}
	return p.s[p.pos:p.pos+len(prefix)] == prefix
}

func (p *RustTypesParser) Consume(prefix string) bool {
	if p.Matches(prefix) {
		p.pos += len(prefix)
		return true
	}
	return false
}

func (p *RustTypesParser) Skip(n int) {
	p.pos += n
	if p.pos > len(p.s) {
		p.pos = len(p.s)
	}
}

func (p *RustTypesParser) skipWhitespace() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func (p *RustTypesParser) parseIdent() string {
	start := p.pos
	for p.pos < len(p.s) {
		r := rune(p.s[p.pos])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			p.pos++
		} else {
			break
		}
	}
	return p.s[start:p.pos]
}

func (p *RustTypesParser) Parse() (RustType, error) {
	p.skipWhitespace()
	return p.parseType()
}

func (p *RustTypesParser) parseType() (RustType, error) {
	p.skipWhitespace()

	// Check for tuple syntax: (T, U, ...)
	if p.Peek() == '(' {
		return p.parseTuple()
	}

	// Check for array syntax: [T; N]
	if p.Peek() == '[' {
		return p.parseArray()
	}

	// Parse path (sequence of identifiers separated by ::)
	var segments []string

	ident := p.parseIdent()
	if ident == "" {
		return RustType{}, fmt.Errorf("expected identifier at position %d", p.pos)
	}
	segments = append(segments, ident)

	p.skipWhitespace()

	// Parse additional path segments
	for p.Consume("::") {
		p.skipWhitespace()

		ident := p.parseIdent()
		if ident == "" {
			return RustType{}, fmt.Errorf("expected identifier after :: at position %d", p.pos)
		}
		segments = append(segments, ident)
		p.skipWhitespace()
	}

	// Check for generics
	if p.Peek() == '<' {
		p.Advance() // consume '<'
		p.skipWhitespace()

		params, err := p.parseGenericParams()
		if err != nil {
			return RustType{}, err
		}

		p.skipWhitespace()
		if p.Peek() != '>' {
			return RustType{}, fmt.Errorf("expected '>' at position %d", p.pos)
		}
		p.Advance() // consume '>'

		return Base(segments, params), nil
	}

	// Return simple path
	return Base(segments, nil), nil
}

func (p *RustTypesParser) parseGenericParams() ([]RustType, error) {
	var params []RustType

	for {
		param, err := p.parseType()
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		p.skipWhitespace()

		if p.Peek() == ',' {
			p.Advance() // consume ','
			p.skipWhitespace()
		} else {
			break
		}
	}

	return params, nil
}

func (p *RustTypesParser) parseTuple() (RustType, error) {
	if p.Peek() != '(' {
		return RustType{}, fmt.Errorf("expected '(' at position %d", p.pos)
	}
	p.Advance() // consume '('
	p.skipWhitespace()

	// Handle empty tuple
	if p.Peek() == ')' {
		p.Advance() // consume ')'
		return Tuple([]RustType{}), nil
	}

	var types []RustType
	for {
		typ, err := p.parseType()
		if err != nil {
			return RustType{}, err
		}
		types = append(types, typ)

		p.skipWhitespace()

		if p.Peek() == ',' {
			p.Advance() // consume ','
			p.skipWhitespace()
		} else {
			break
		}
	}

	p.skipWhitespace()
	if p.Peek() != ')' {
		return RustType{}, fmt.Errorf("expected ')' at position %d", p.pos)
	}
	p.Advance() // consume ')'

	return Tuple(types), nil
}

func (p *RustTypesParser) parseArray() (RustType, error) {
	if p.Peek() != '[' {
		return RustType{}, fmt.Errorf("expected '[' at position %d", p.pos)
	}
	p.Advance() // consume '['
	p.skipWhitespace()

	base, err := p.parseType()
	if err != nil {
		return RustType{}, err
	}

	p.skipWhitespace()
	if p.Peek() != ';' {
		return RustType{}, fmt.Errorf("expected ';' at position %d", p.pos)
	}
	p.Advance() // consume ';'
	p.skipWhitespace()

	// Parse array length (simple integer)
	len := p.parseInt()
	if len < 0 {
		return RustType{}, fmt.Errorf("expected array length at position %d", p.pos)
	}

	p.skipWhitespace()
	if p.Peek() != ']' {
		return RustType{}, fmt.Errorf("expected ']' at position %d", p.pos)
	}
	p.Advance() // consume ']'

	return Array(base, len), nil
}

func (p *RustTypesParser) parseInt() int {
	start := p.pos
	for p.pos < len(p.s) && unicode.IsDigit(rune(p.s[p.pos])) {
		p.pos++
	}
	if start == p.pos {
		return -1
	}

	result := 0
	for i := start; i < p.pos; i++ {
		result = result*10 + int(p.s[i]-'0')
	}
	return result
}
