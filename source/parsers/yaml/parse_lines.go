package yaml

import "strings"

func parse_lines(input string) []parser_line {

	lines  := strings.Split(input, "\n")
	result := make([]parser_line, 0)

	block_indent := -1

	for index, line := range lines {

		indent  := lineIndent(line)
		trimmed := strings.TrimRight(line, " \t")

		if block_indent >= 0 {

			if strings.TrimSpace(trimmed) == "" {
				continue
			}

			if indent > block_indent {

				result = append(result, parser_line{
					Number: index + 1,
					Indent: indent,
					Text:   strings.TrimSpace(trimmed),
				})

				continue

			}

			block_indent = -1

		}

		text := strings.TrimSpace(trimmed)

		if text == "" {
			continue
		}

		if strings.HasPrefix(text, "#") {
			continue
		}

		text = stripInlineComment(text)

		if strings.TrimSpace(text) == "" {
			continue
		}

		result = append(result, parser_line{
			Number: index + 1,
			Indent: indent,
			Text:   text,
		})

		if isBlockScalar(text) {
			block_indent = indent
		}

	}

	return result

}

func lineIndent(line string) int {

	indent := 0

	for _, character := range line {

		if character == ' ' || character == '\t' {
			indent++
			continue
		}

		break

	}

	return indent

}

func isBlockScalar(text string) bool {

	return strings.HasSuffix(text, ": |") ||
		strings.HasSuffix(text, ": |-") ||
		strings.HasSuffix(text, ": |+")

}

// stripInlineComment removes a trailing "# ..." comment from a line. It only
// treats "#" as a comment when it starts the value or is preceded by
// whitespace, and it ignores hashes inside single or double quoted strings.
func stripInlineComment(text string) string {

	in_single := false
	in_double := false
	escaped   := false

	for index := 0; index < len(text); index++ {

		chr := text[index]

		if escaped == true {
			escaped = false
			continue
		}

		if chr == '\\' && in_double == true {
			escaped = true
			continue
		}

		if chr == '\'' && in_double == false {
			in_single = !in_single
			continue
		}

		if chr == '"' && in_single == false {
			in_double = !in_double
			continue
		}

		if chr == '#' && in_single == false && in_double == false {

			if index == 0 || text[index-1] == ' ' || text[index-1] == '\t' {
				return strings.TrimRight(text[0:index], " \t")
			}

		}

	}

	return strings.TrimRight(text, " \t")

}
