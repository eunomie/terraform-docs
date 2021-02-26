/*
Copyright 2021 The terraform-docs Authors.

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/

package format

import (
	"fmt"
	"regexp"
	"strings"
)

// sanitize cleans a Markdown document to soothe linters.
func sanitize(markdown string) string {
	result := markdown

	// Preserve double spaces at the end of the line
	result = regexp.MustCompile(` {2}(\r?\n)`).ReplaceAllString(result, "‡‡‡DOUBLESPACES‡‡‡$1")

	// Remove trailing spaces from the end of lines
	result = regexp.MustCompile(` +(\r?\n)`).ReplaceAllString(result, "$1")
	result = regexp.MustCompile(` +$`).ReplaceAllLiteralString(result, "")

	// Preserve double spaces at the end of the line
	result = regexp.MustCompile(`‡‡‡DOUBLESPACES‡‡‡(\r?\n)`).ReplaceAllString(result, "  $1")

	// Remove blank line with only double spaces in it
	result = regexp.MustCompile(`(\r?\n)  (\r?\n)`).ReplaceAllString(result, "$1")

	// Remove multiple consecutive blank lines
	result = regexp.MustCompile(`(\r?\n){3,}`).ReplaceAllString(result, "$1$1")
	result = regexp.MustCompile(`(\r?\n){2,}$`).ReplaceAllString(result, "")

	return result
}

// printFencedCodeBlock prints codes in fences, it automatically detects if
// the input 'code' contains '\n' it will use multi line fence, otherwise it
// wraps the 'code' inside single-tick block.
// If the fenced is multi-line it also appens an extra '\n` at the end and
// returns true accordingly, otherwise returns false for non-carriage return.
func printFencedCodeBlock(code string, language string) (string, bool) {
	if strings.Contains(code, "\n") {
		return fmt.Sprintf("\n\n```%s\n%s\n```\n", language, code), true
	}
	return fmt.Sprintf("`%s`", code), false
}

func printFencedCodeBlockWithIndent(code string, language string) string {
	if !strings.Contains(code, "\n") {
		return fmt.Sprintf("`%s`", code)
	}
	block := fmt.Sprintf("```%s\n%s\n```", language, code)
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		lines[i] = "    " + l
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}

// printFencedAsciidocCodeBlock prints codes in fences, it automatically detects if
// the input 'code' contains '\n' it will use multi line fence, otherwise it
// wraps the 'code' inside single-tick block.
// If the fenced is multi-line it also appens an extra '\n` at the end and
// returns true accordingly, otherwise returns false for non-carriage return.
func printFencedAsciidocCodeBlock(code string, language string) (string, bool) {
	if strings.Contains(code, "\n") {
		return fmt.Sprintf("\n[source,%s]\n----\n%s\n----\n", language, code), true
	}
	return fmt.Sprintf("`%s`", code), false
}

// convertMultiLineText converts a multi-line text into a suitable Markdown representation.
func convertMultiLineText(s string, isTable bool) string {
	if isTable {
		s = strings.TrimSpace(s)
	}

	// Convert double newlines to <br><br>.
	s = strings.Replace(s, "\n\n", "<br><br>", -1)

	// Convert line-break on a non-empty line followed by another line
	// starting with "alphanumeric" word into space-space-newline
	// which is a know convention of Markdown for multi-lines paragprah.
	// This doesn't apply on a markdown list for example, because all the
	// consecutive lines start with hyphen which is a special character.
	s = regexp.MustCompile(`(\S*)(\r?\n)(\s*)(\w+)`).ReplaceAllString(s, "$1  $2$3$4")
	s = strings.Replace(s, "    \n", "  \n", -1)
	s = strings.Replace(s, "<br>  \n", "\n\n", -1)

	if isTable {
		// Convert space-space-newline to <br>
		s = strings.Replace(s, "  \n", "<br>", -1)

		// Convert single newline to <br>.
		s = strings.Replace(s, "\n", "<br>", -1)
	} else {
		s = strings.Replace(s, "<br>", "\n", -1)
	}

	return s
}
