package peg

import "unicode/utf8"

const (
	asciinl    = 10
	asciicr    = 12
	asciispace = 32
	ascii0     = 48
	ascii9     = 57
	asciiA     = 65
	asciiZ     = 90
	asciia     = 97
	asciiz     = 122
)

func isNewline(char rune) bool {
	return char == asciinl || char == asciicr
}

func isSpace(char rune) bool {
	return char == asciispace
}

func isLetter(char rune) bool {
	return (asciiA <= char && char <= asciiZ) || (asciia <= char && char <= asciiz)
}

func isDigit(char rune) bool {
	return ascii0 <= char && char <= ascii9
}

func lenRune(r rune) int {
	if 0x0 <= r && r <= 0x7f {
		return 1
	} else if 0x80 <= r && r <= 0x7ff {
		return 2
	} else if 0x800 <= r && r <= 0xffff {
		return 3
	} else {
		return 4
	}
}

func byteToRune(b []byte) []rune {
	runes := []rune{}
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)
		b = b[size:]
	}
	return runes
}

// TODO finish escape
func escape(char rune) rune {
	switch char {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	}
	return char
}
