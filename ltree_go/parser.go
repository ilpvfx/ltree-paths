package ltree_go

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type LTreePathErrorCode uint

const (
	ExpectedLabel LTreePathErrorCode = iota
	InvalidCharacter
	LabelTooLong
	PathTooLong
	// Go strings can contain whatever bytes the user desires, and this might give us problems, for
	// example:
	// - lone surrogate halves
	// - broken continuation sequences in UTF-8
	BadUnicode
)

func makeError(code LTreePathErrorCode, offset uint) *LTreePathError {
	msg := ""
	switch code {
	case ExpectedLabel:
		msg = fmt.Sprintf("Parse error: Expected label at offset %v", offset)
	case InvalidCharacter:
		msg = fmt.Sprintf("Parse error: Invalid character at offset %v", offset)
	case LabelTooLong:
		msg = fmt.Sprintf("Parse error: Label too long (max %v UTF-8 scalar values)", MAX_LABEL_USV_COUNT)
	case PathTooLong:
		msg = fmt.Sprintf("Parse error: Path too long (max %v labels)", MAX_PATH_LABEL_COUNT)
	case BadUnicode:
		msg = fmt.Sprintf("Parse error: Bad UTF-8 sequence at offset %v", offset)
	default:
		panic("Non exhaustive switch")
	}
	return &LTreePathError{msg, code}
}

func runeAllowedInLabels(r rune) bool {
	const charMinus rune = '-'
	const charUnderscore rune = '_'

	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == charMinus || r == charUnderscore
}

// Iterate over the byte slices, each slice representing the string separated by CharDelimiter
//
// Iteration continues until the callback gives an error or the last string is returned.
//
// Note that a string such as ".abc." generates "", "abc", "".
func sliceLabelsFunc(s string, callback func(s string) error) error {
	delimiter := string(CharDelimiter)
	last, rem, found := strings.Cut(s, delimiter)
	for found {
		if err := callback(last); err != nil {
			return err
		}
		last, rem, found = strings.Cut(rem, delimiter)
	}

	if err := callback(last); err != nil {
		return err
	}

	return nil
}

// Decodes the UTF-8 and checks if it is a valid label. Returns the new USV count and any potential
// error.
func validateLabel(s string, usvCount uint) (uint, error) {
	if len(s) == 0 {
		return usvCount, makeError(ExpectedLabel, usvCount)
	}

	bytes := []byte(s)
	var labelUsvCount uint = 0

	for len(bytes) > 0 {
		r, width := utf8.DecodeRune(bytes)
		if r == utf8.RuneError {
			return usvCount, makeError(BadUnicode, usvCount)
		}

		if runeAllowedInLabels(r) {
			usvCount += 1
			labelUsvCount += 1

			if labelUsvCount > MAX_LABEL_USV_COUNT {
				return usvCount, makeError(LabelTooLong, usvCount)
			}
		} else {
			return usvCount, makeError(InvalidCharacter, usvCount)
		}

		bytes = bytes[width:]
	}

	return usvCount, nil
}

func validate(s string) error {
	if len(s) == 0 {
		return nil
	}

	var labelCount uint = 0
	var usvCount uint = 0

	err := sliceLabelsFunc(s, func(s string) error {
		if labelCount > 0 {
			// We add 1 to represent the "." separating the labels.
			usvCount += 1
		}

		newUsvCount, err := validateLabel(s, usvCount)
		if err != nil {
			return err
		}
		usvCount = newUsvCount

		labelCount += 1
		if labelCount > MAX_PATH_LABEL_COUNT {
			return makeError(PathTooLong, usvCount)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if usvCount == 0 && labelCount == 0 {
		return makeError(ExpectedLabel, usvCount)
	}

	return nil
}
