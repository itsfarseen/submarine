package rust_types

import (
	"fmt"
	"strings"
	"unicode"
)

type RustType interface {
	String() string
}

type Ident struct {
	Name string
}

func (i Ident) String() string {
	return i.Name
}

type Generic struct {
	Outer  string
	Params []RustType
}

func (g Generic) String() string {
	if len(g.Params) == 0 {
		return g.Outer
	}

	var params []string
	for _, param := range g.Params {
		params = append(params, param.String())
	}

	return fmt.Sprintf("%s<%s>", g.Outer, strings.Join(params, ", "))
}

type Assoc struct {
	Outer RustType
	Next  RustType
}

func (a Assoc) String() string {
	return fmt.Sprintf("%s::%s", a.Outer.String(), a.Next.String())
}

type AsTrait struct {
	Src    RustType
	Target RustType
}

func (at AsTrait) String() string {
	return fmt.Sprintf("<%s as %s>", at.Src.String(), at.Target.String())
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

	// Check for AsTrait syntax: <Foo as Bar>
	if p.Peek() == '<' {
		return p.parseAsTrait()
	}

	// Parse basic type (identifier or generic)
	typ, err := p.parseBasicType()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()

	// Check for association (::)
	if p.Matches("::") {
		return p.parseAssoc(typ)
	}

	return typ, nil
}

func (p *RustTypesParser) parseBasicType() (RustType, error) {
	p.skipWhitespace()

	ident := p.parseIdent()
	if ident == "" {
		return nil, fmt.Errorf("expected identifier at position %d", p.pos)
	}

	p.skipWhitespace()

	// Check for generic syntax
	if p.Peek() == '<' {
		p.Advance() // consume '<'
		p.skipWhitespace()

		params, err := p.parseGenericParams()
		if err != nil {
			return nil, err
		}

		p.skipWhitespace()
		if p.Peek() != '>' {
			return nil, fmt.Errorf("expected '>' at position %d", p.pos)
		}
		p.Advance() // consume '>'

		return Generic{Outer: ident, Params: params}, nil
	}

	return Ident{Name: ident}, nil
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

func (p *RustTypesParser) parseAssoc(outer RustType) (RustType, error) {
	for p.Matches("::") {
		p.Skip(2) // consume "::"
		p.skipWhitespace()

		next, err := p.parseBasicType()
		if err != nil {
			return nil, err
		}

		outer = Assoc{Outer: outer, Next: next}
		p.skipWhitespace()
	}

	return outer, nil
}

func (p *RustTypesParser) parseAsTrait() (RustType, error) {
	if p.Peek() != '<' {
		return nil, fmt.Errorf("expected '<' at position %d", p.pos)
	}
	p.Advance() // consume '<'
	p.skipWhitespace()

	src, err := p.parseType()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if !p.Matches("as") {
		return nil, fmt.Errorf("expected 'as' at position %d", p.pos)
	}
	p.Skip(2) // consume "as"
	p.skipWhitespace()

	target, err := p.parseTypeInAsTrait()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.Peek() != '>' {
		return nil, fmt.Errorf("expected '>' at position %d", p.pos)
	}
	p.Advance() // consume '>'

	return AsTrait{Src: src, Target: target}, nil
}

func (p *RustTypesParser) parseTypeInAsTrait() (RustType, error) {
	p.skipWhitespace()

	// Parse basic type (identifier or generic) - but no AsTrait recursion
	typ, err := p.parseBasicType()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()

	// Check for association (::)
	if p.Matches("::") {
		return p.parseAssoc(typ)
	}

	return typ, nil
}
