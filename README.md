# PEG.go

A parser generator written in Go.

Use [Parsing Expression Grammar](https://en.wikipedia.org/wiki/Parsing_expression_grammar).

- [Installation](#installation)
- [Usage](#usage)
  - [Example](#example)
  - [Generating a Parser](#generating-a-parser)
  - [Using the Parser](#using-the-parser)
- [Syntax](#syntax)

## Installation

```
$ go get github.com/laurence6/PEG.go
```

## Usage

### Example

```
# 1.peg
```

### Generating a Parser

```go
// main.go
package main

import (
	"io"
	"os"
)

import "github.com/laurence6/PEG.go"

func main() {
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	peg.GenerateParser(in, out)
}
```
```
$ go run main.go < 1.peg > 1.go
```

### Using the Parser

```
```

## Syntax

- `package xxx`

  Like package declaration in Go

- `import " "`

  `import . " "`

  `import " " " "`

  Like import declaration in Go

- `"a \n b"`

  Literal string

- `[_a-z0-9]`

  Character class

- `.`

  Any character

- `(e)`

  Grouping

- `e?`

  Zero or One

- `e*`

  Zero or More

- `e+`

  One or More

- `&e`

  And-predicate

- `!e`

  Not-predicate

- `e1 e2`

  Sequence

- `e1 / e2`

  Prioritized choice

## License

Copyright (C) 2016-2017  Laurence Liu <liuxy6@gmail.com>

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program.  If not, see <http://www.gnu.org/licenses/>.
