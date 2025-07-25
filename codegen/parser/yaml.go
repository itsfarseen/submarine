package parser

import (
	"bufio"
	"io"
	"strings"
)

type YamlObject struct {
	Fields []*YamlKV
}

type YamlKV struct {
	Key   string
	Value any
}

type YamlArray struct {
	Items []any
}

type YamlParser struct {
	lines []string
	index int
}

func (p *YamlParser) Peek() string {
	if p.IsEOF() {
		return ""
	}
	return p.lines[p.index]
}

func (p *YamlParser) Advance() {
	if !p.IsEOF() {
		p.index++
	}
}

func (p *YamlParser) IsEOF() bool {
	return p.index >= len(p.lines)
}

func ParseYAML(reader io.Reader) (any, error) {
	scanner := bufio.NewScanner(reader)
	var lines []string

	// Read all lines
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if strings.TrimSpace(line) == "" || strings.TrimSpace(line)[0] == '#' {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, nil
	}

	parser := &YamlParser{lines: lines, index: 0}
	return parser.parseValue(0), nil
}

func (p *YamlParser) parseValue(expectedIndent int) any {
	if p.IsEOF() {
		return nil
	}

	line := p.Peek()
	indent := CountIndent(line)

	// If indentation is less than expected, stop parsing at this level
	if indent < expectedIndent {
		return nil
	}

	trimmed := strings.TrimSpace(line)

	// Check if this is an array item
	if strings.HasPrefix(trimmed, "- ") {
		return p.parseArray(indent)
	}

	// Check if this is a key-value pair
	if strings.Contains(trimmed, ":") {
		return p.parseObject(indent)
	}

	// It's a simple string value
	p.Advance()
	return trimmed
}

func (p *YamlParser) parseArray(baseIndent int) *YamlArray {
	array := &YamlArray{}

	for !p.IsEOF() {
		line := p.Peek()
		indent := CountIndent(line)
		trimmed := strings.TrimSpace(line)

		// If indent is less than base, we're done with this array
		if indent < baseIndent {
			break
		}

		// If it's not an array item at this level, we're done
		if indent == baseIndent && !strings.HasPrefix(trimmed, "- ") {
			break
		}

		if strings.HasPrefix(trimmed, "- ") {
			itemContent := strings.TrimSpace(trimmed[2:]) // Remove "- "
			p.Advance()

			if itemContent == "" {
				// Multi-line array item, parse the next lines as an object or value
				value := p.parseValue(indent + 2)
				array.Items = append(array.Items, value)
			} else if strings.Contains(itemContent, ":") {
				// Inline object - but we need to check if there are more fields following
				obj := &YamlObject{}
				parts := strings.SplitN(itemContent, ":", 2)
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				obj.Fields = append(obj.Fields, &YamlKV{Key: key, Value: val})

				// Check if there are more fields at the same indentation level for this array item
				for !p.IsEOF() {
					nextLine := p.Peek()
					nextIndent := CountIndent(nextLine)
					nextTrimmed := strings.TrimSpace(nextLine)

					// If we encounter another array item or lower indentation, stop
					if nextIndent < indent+2 || strings.HasPrefix(nextTrimmed, "- ") {
						break
					}

					// If it's at the right indentation and contains a colon, it's part of this object
					if nextIndent == indent+2 && strings.Contains(nextTrimmed, ":") {
						parts := strings.SplitN(nextTrimmed, ":", 2)
						key := strings.TrimSpace(parts[0])
						val := strings.TrimSpace(parts[1])
						obj.Fields = append(obj.Fields, &YamlKV{Key: key, Value: val})
						p.Advance()
					} else {
						break
					}
				}

				array.Items = append(array.Items, obj)
			} else {
				// Simple string item
				array.Items = append(array.Items, itemContent)
			}
		} else {
			break
		}
	}

	return array
}

func (p *YamlParser) parseObject(baseIndent int) *YamlObject {
	obj := &YamlObject{}

	for !p.IsEOF() {
		line := p.Peek()
		indent := CountIndent(line)
		trimmed := strings.TrimSpace(line)

		// If indent is less than base, we're done with this object
		if indent < baseIndent {
			break
		}

		// If this is an array item, we're done
		if strings.HasPrefix(trimmed, "- ") {
			break
		}

		// Must be a key-value pair at this level
		if indent == baseIndent && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			key := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])
			p.Advance()

			if valueStr == "" {
				// Multi-line value, parse nested content
				value := p.parseValue(indent + 2)
				obj.Fields = append(obj.Fields, &YamlKV{Key: key, Value: value})
			} else {
				// Inline value
				obj.Fields = append(obj.Fields, &YamlKV{Key: key, Value: valueStr})
			}
		} else if indent > baseIndent {
			// This should be handled by recursive calls
			break
		} else {
			p.Advance()
		}
	}

	return obj
}

func CountIndent(s string) int {
	count := 0
	for _, char := range s {
		if char == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}
