package bencodeParser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBencodeParser(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{"i42e", 42},                            // Integer
		{"4:test", "test"},                      // String
		{"l3:onei2ee", []interface{}{"one", 2}}, // List
		{"d3:fooi42e3:bar3:baze", map[string]interface{}{"foo": 42, "bar": "baz"}}, // Dictionary
		{"li42e4:teste", []interface{}{42, "test"}},                                // Mixed List
		{"d4:dictd3:fooi42e3:bar3:baz4:listli1ei2e5:threeeee", map[string]interface{}{
			"dict": map[string]interface{}{"foo": 42, "bar": "baz", "list": []interface{}{1, 2, "three"}},
		}}, // Nested Dictionary and List

		{"le", []interface{}{}},          // Empty List
		{"de", map[string]interface{}{}}, // Empty Dictionary
	}

	for _, testCase := range testCases {
		parser := NewBencodeParser([]byte(testCase.input))
		result, err := parser.Parse()
		if err != nil {
			t.Errorf("Error parsing Bencode: %v", err)
		}

		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("For input '%s', expected %+v, got %+v", testCase.input, testCase.expected, result)
		}
	}

	bto, err := OpenTorrent("D:\\Programming stuff\\Projects\\Go\\bitTorrent Client\\torrentfile\\archlinux-2019.12.01-x86_64.iso.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v\n%v\n%v\n%v\n", bto.Announce, bto.Info.Name, bto.Info.Length, bto.Info.PieceLength)

}
