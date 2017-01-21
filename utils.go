package peg

import (
	"io"
	"os"
)

var BufferSize = 200

func GenerateParser(r io.Reader, w io.Writer) {
	// New scanner with buffer size
	scanner := NewScanner(r, BufferSize)

	// Get all tokens
	tokens := []*Token{}
	for {
		tok := scanner.Scan()
		tokens = append(tokens, &tok)
		if tok.Type == EOF {
			break
		}
	}

	// Get AST from tokens
	tree := GetTree(tokens)

	// Generate parser
	tree.GenCode(os.Stdout)
}
