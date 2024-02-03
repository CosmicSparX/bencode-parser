package bencodeParser

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode/utf8"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type BencodeParser struct {
	encoding []rune
	lexer    BencodeLexer
}

func NewBencodeParser(input []rune) *BencodeParser {
	return &BencodeParser{encoding: input, lexer: BencodeLexer{input: input}}
}

func Open(path string) (BencodeTorrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return BencodeTorrent{}, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fileSize := fileInfo.Size()

	// Read the content of the torrent file
	content := make([]byte, fileSize)
	_, err = file.Read(content)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(content), utf8.RuneCount(content))

	input := bytes.Runes(content)

	fmt.Println(string(input[269+39921]))
	parser := NewBencodeParser(input)
	raw, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(raw)

	return BencodeTorrent{}, nil
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
