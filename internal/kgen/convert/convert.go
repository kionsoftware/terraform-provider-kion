// Package convert provides name conversion utilities for code generation.
package convert

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ToSnakeCase converts a PascalCase string to snake_case.
// Copied from github.com/hashicorp/terraform-provider-aws/names.ToSnakeCase
func ToSnakeCase(in string) string {
	out := strings.Builder{}

	for i, ch := range []byte(in) {
		isCap := isCapitalLetter(ch)
		isLow := isLowercaseLetter(ch)
		isDig := isNumeric(ch)

		if isCap {
			ch = toLowercaseLetter(ch)
		}

		if i < len(in)-1 {
			nextCh := in[i+1]
			nextIsCap := isCapitalLetter(nextCh)
			nextIsLow := isLowercaseLetter(nextCh)
			nextIsDig := isNumeric(nextCh)

			// Append underscore if case changes.
			if (isCap && nextIsLow) || (isLow && (nextIsCap || nextIsDig) || (isDig && (nextIsCap || nextIsLow))) {
				if isCap && nextIsLow {
					if prevIsCap := i > 0 && isCapitalLetter(in[i-1]); prevIsCap {
						out.WriteByte('_')
					}
				}
				out.WriteByte(ch)
				if isLow || isDig {
					out.WriteByte('_')
				}

				continue
			}
		}

		if isCap || isLow || isDig {
			out.WriteByte(ch)
		} else {
			out.WriteByte('_')
		}
	}

	return out.String()
}

func isCapitalLetter(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}

func isLowercaseLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}

func isNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func toLowercaseLetter(ch byte) byte {
	ch += 'a'
	ch -= 'A'
	return ch
}

// ToHumanResName converts a camel cased string to a human readable name
func ToHumanResName(upper string) string {
	re := regexp.MustCompile(`([a-z])([A-Z]{2,})`)
	upper = re.ReplaceAllString(upper, `${1} ${2}`)

	re2 := regexp.MustCompile(`([A-Z][a-z])`)
	return strings.TrimPrefix(re2.ReplaceAllString(upper, ` $1`), " ")
}

// ToProviderResourceName adds the kion_ prefix to a snake cased name
// of a resource or data source
func ToProviderResourceName(_, snakeName string) string {
	return fmt.Sprintf("kion_%s", snakeName)
}

// ToLowercasePrefix converts a string beginning with uppercase letters
// to begin lowercased
//
// Specifically, this is used to take user-provided input beginning with uppercase
// letters, and transform it so that it can be used to name a private struct. This
// function assumes all characters are unicode.
func ToLowercasePrefix(s string) string {
	var hasLower bool
	var splitIdx int
	for i, char := range s {
		if unicode.IsLower(char) {
			hasLower = true
			break
		}
		splitIdx = i
	}

	if !hasLower {
		return strings.ToLower(s)
	}
	if splitIdx == 0 && len(s) > 0 {
		splitIdx++
	}
	return strings.ToLower(s[:splitIdx]) + s[splitIdx:]
}
