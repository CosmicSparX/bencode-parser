package bencodeParser

import (
	"errors"
	"strconv"
	"unicode"
)

type TokenType int

const (
	Integer TokenType = iota
	String
	ListStart
	DictStart
	End
)

type Token struct {
	Type  TokenType
	Value string
}

type BencodeLexer struct {
	input   []rune
	current int
}

func NewBencodeLexer(input string) *BencodeLexer {
	return &BencodeLexer{input: []rune(input)}
}

func (l *BencodeLexer) NextToken() (Token, error) {
	for l.current < len(l.input) && unicode.IsSpace(l.input[l.current]) {
		l.current++
	}

	if l.current >= len(l.input) {
		return Token{Type: End, Value: ""}, nil
	}

	currentChar := l.input[l.current]

	if unicode.IsDigit(currentChar) {
		return l.readString()
	} else if currentChar == 'i' {
		l.current++
		integerToken := l.readInteger()
		l.current++
		return integerToken, nil
	} else if currentChar == 'l' {
		l.current++
		return Token{Type: ListStart, Value: "l"}, nil
	} else if currentChar == 'd' {
		l.current++
		return Token{Type: DictStart, Value: "d"}, nil
	} else if currentChar == 'e' {
		l.current++
		return Token{Type: End, Value: "e"}, nil
	} else {
		return Token{Type: End, Value: ""}, errors.New("invalid Character")
	}
}

func (l *BencodeLexer) readInteger() Token {
	start := l.current

	if l.input[l.current] == '-' {
		l.current++
	}

	for l.current < len(l.input) && unicode.IsDigit(l.input[l.current]) {
		l.current++
	}
	end := l.current
	return Token{Type: Integer, Value: string(l.input[start:end])}
}

func (l *BencodeLexer) readString() (Token, error) {
	length, _ := strconv.Atoi(l.readInteger().Value)
	start := l.current + 1

	if l.current+length > len(l.input) {
		return Token{Type: End, Value: ""}, errors.New("wrong length for string token")
	}
	end := start + length
	l.current += length + 1
	return Token{Type: String, Value: string(l.input[start:end])}, nil
}
