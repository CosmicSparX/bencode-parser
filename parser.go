package bencodeParser

import (
	"io"
	"log"
	"os"
	"strconv"
)

type BencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     BencodeInfo `bencode:"info"`
}

type BencodeParser struct {
	encoding []byte
	lexer    BencodeLexer
}

func NewBencodeParser(input []byte) *BencodeParser {
	return &BencodeParser{encoding: input, lexer: BencodeLexer{input: input}}
}

func OpenTorrent(path string) (BencodeTorrent, error) {
	file, err := os.Open(path)
	if err != nil {
		return BencodeTorrent{}, err
	}

	input, _ := io.ReadAll(file)

	parser := NewBencodeParser(input)
	raw, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	data := raw.(map[string]interface{})
	info := data["info"].(map[string]interface{})

	bto := BencodeTorrent{
		Announce: data["announce"].(string),
		Info: BencodeInfo{
			Length:      info["length"].(int),
			Name:        info["name"].(string),
			PieceLength: info["piece length"].(int),
			Pieces:      info["pieces"].(string),
		},
	}

	return bto, nil
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
