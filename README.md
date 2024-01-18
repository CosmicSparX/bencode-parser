# Bencode Parser

Usage:

```go
package main

import (
	"fmt"
	"github.com/CosmicSparX/bencode-parser"
)

func main() {
	input := "your Bencode-encoded data here"
	parser := bencodeParser.NewBencodeParser(input)

	result, err := parser.Parse()
	if err != nil {
		fmt.Println("Error parsing Bencode:", err)
		return
	}

	fmt.Println("Parsed result:", result)
}

```
