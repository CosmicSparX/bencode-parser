package bencodeParser

import (
	"strconv"
)

type BencodeParser struct {
	encoding string
	lexer    BencodeLexer
}

func NewBencodeParser(input string) *BencodeParser {
	return &BencodeParser{encoding: input, lexer: BencodeLexer{input: []rune(input)}}
}

func (p *BencodeParser) Parse() (interface{}, error) {
	token, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case Integer:
		return strconv.Atoi(token.Value)
	case String:
		return token.Value, nil
	case ListStart:
		return p.parseList()
	case DictStart:
		return p.parseDict()
	case End:
		return nil, nil
	default:
		panic("unhandled default case")
	}
}

func (p *BencodeParser) parseList() (interface{}, error) {
	list := make([]interface{}, 0)
	for {
		item, err := p.Parse()
		if err != nil {
			return nil, err
		}

		if item == nil {
			break
		}

		list = append(list, item)
	}
	return list, nil
}

func (p *BencodeParser) parseDict() (interface{}, error) {
	dict := make(map[string]interface{})

	for {
		key, err := p.lexer.NextToken()
		if err != nil {
			return nil, err
		}

		if key.Type == End {
			break
		}

		value, err := p.Parse()
		if err != nil {
			return nil, err
		}
		dict[key.Value] = value
	}
	return dict, nil
}
