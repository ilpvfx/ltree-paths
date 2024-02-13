package ltree_go

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ParseErrorCode uint

const (
	ExpectedLabel ParseErrorCode = iota
	InvalidCharacter
	LabelTooLong
	PathTooLong
	// Go strings can contain whatever bytes the user desires, and this might give us problems, for
	// example:
	// - lone surrogate halves
	// - broken continuation sequences in UTF-8
	BadUnicode
)

func makeError(code ParseErrorCode, offset uint) *LTreePathError {
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

type parsingState struct {
	usvOffset    uint
	bytes        []byte
	currentLabel currentLabel
}

func (p *parsingState) ErrorAtCurrentOffset(what ParseErrorCode) error {
	return makeError(what, p.usvOffset)
}

func (p *parsingState) BeginLabel() {
	p.currentLabel = currentLabel{}
}

func (p *parsingState) Advance(runeWidth int) {
	p.bytes = p.bytes[runeWidth:]
	p.usvOffset += 1
}

func (p *parsingState) IsCurrentLabelEmpty() bool {
	return p.currentLabel.usvCount == 0
}

type currentLabel struct {
	usvCount uint
}

func runeAllowedInLabels(r rune) bool {
	const charMinus rune = '-'
	const charUnderscore rune = '_'

	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == charMinus || r == charUnderscore
}

// Iterate over the byte slices, each slice representing the string separated by CharDelimiter
//
// Iteration continues until the callback gives an error or the last string is returned.
func sliceLabelsFunc(s string, callback func(s string) error) error {
	delimiter := string(CharDelimiter)
	last, rem, found := strings.Cut(s, delimiter)
	for found {
		if len(last) > 0 {
			if err := callback(last); err != nil {
				return err
			}
		}
		last, rem, found = strings.Cut(rem, delimiter)
	}

	if len(last) > 0 {
		if err := callback(last); err != nil {
			return err
		}
	}

	return nil
}

func validate(s string) error {
	// TODO (siot): with a small application of some strategy function we could make this adaptible
	// for the builder later when we want to turn a path into a builder.

	if len(s) == 0 {
		return nil
	}

	var labelCount uint = 0
	p := &parsingState{usvOffset: 0, bytes: []byte(s)}
	p.BeginLabel()

	for len(p.bytes) > 0 {
		r, width := utf8.DecodeRune(p.bytes)
		if r == utf8.RuneError {
			return p.ErrorAtCurrentOffset(BadUnicode)
		}

		if runeAllowedInLabels(r) {
			p.currentLabel.usvCount += 1

			if p.currentLabel.usvCount > MAX_LABEL_USV_COUNT {
				return makeError(LabelTooLong, p.usvOffset+1)
			}
		} else if r == CharDelimiter {
			if p.IsCurrentLabelEmpty() {
				return p.ErrorAtCurrentOffset(ExpectedLabel)
			}

			labelCount += 1

			if labelCount > MAX_PATH_LABEL_COUNT {
				return p.ErrorAtCurrentOffset(PathTooLong)
			}

			p.BeginLabel()
		} else {
			return p.ErrorAtCurrentOffset(InvalidCharacter)
		}

		p.Advance(width)
	}

	// The last label will not end with a delimiter. Thus, we'll have to special case it.
	if p.IsCurrentLabelEmpty() {
		return p.ErrorAtCurrentOffset(ExpectedLabel)
	} else {
		labelCount += 1
	}

	if labelCount > MAX_PATH_LABEL_COUNT {
		return p.ErrorAtCurrentOffset(PathTooLong)
	}

	return nil
}
