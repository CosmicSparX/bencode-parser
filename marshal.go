// Copyright 2009 The Go Authors. All rights reserved.
// Taken from github.com/jackpal/bencode-go/blob/master/struct.go

package bencodeParser

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

type MarshalError struct {
	T reflect.Type
}

func (e *MarshalError) Error() string {
	return "bencode cannot encode value of type " + e.T.String()
}

func writeArrayOrSlice(w io.Writer, val reflect.Value) (err error) {
	_, err = fmt.Fprint(w, "l")
	if err != nil {
		return
	}
	for i := 0; i < val.Len(); i++ {
		if err := writeValue(w, val.Index(i)); err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(w, "e")
	if err != nil {
		return
	}
	return nil
}

type stringValue struct {
	key       string
	value     reflect.Value
	omitEmpty bool
}

type stringValueArray []stringValue

// Satisfy sort.Interface

func (a stringValueArray) Len() int { return len(a) }

func (a stringValueArray) Less(i, j int) bool { return a[i].key < a[j].key }

func (a stringValueArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func writeSVList(w io.Writer, svList stringValueArray) (err error) {
	sort.Sort(svList)

	for _, sv := range svList {
		if sv.isValueNil() {
			continue // Skip null values
		}
		s := sv.key
		_, err = fmt.Fprintf(w, "%d:%s", len(s), s)
		if err != nil {
			return
		}

		if err = writeValue(w, sv.value); err != nil {
			return
		}
	}
	return
}

func writeMap(w io.Writer, val reflect.Value) (err error) {
	key := val.Type().Key()
	if key.Kind() != reflect.String {
		return &MarshalError{val.Type()}
	}
	_, err = fmt.Fprint(w, "d")
	if err != nil {
		return
	}

	keys := val.MapKeys()

	// Sort keys

	svList := make(stringValueArray, len(keys))
	for i, key := range keys {
		svList[i].key = key.String()
		svList[i].value = val.MapIndex(key)
	}

	err = writeSVList(w, svList)
	if err != nil {
		return
	}

	_, err = fmt.Fprint(w, "e")
	if err != nil {
		return
	}
	return
}

func bencodeKey(field reflect.StructField, sv *stringValue) (key string) {
	key = field.Name
	tag := field.Tag
	if len(tag) > 0 {
		// Backwards compatability
		// If there's a bencode key/value entry in the tag, use it.
		var tagOpt tagOptions
		key, tagOpt = parseTag(tag.Get("bencode"))
		if len(key) == 0 {
			key = tag.Get("bencode")
			if len(key) == 0 && !strings.Contains(string(tag), ":") {
				// If there is no ":" in the tag, assume it is an old-style tag.
				key = string(tag)
			} else {
				key = field.Name
			}
		}
		if sv != nil && tagOpt.Contains("omitempty") {
			sv.omitEmpty = true
		}
	}
	if sv != nil {
		sv.key = key
	}
	return
}

// tagOptions is the string following a comma in a struct field's "bencode"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's bencode tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, ""
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func writeStruct(w io.Writer, val reflect.Value) (err error) {
	_, err = fmt.Fprint(w, "d")
	if err != nil {
		return
	}

	typ := val.Type()

	numFields := val.NumField()
	svList := make(stringValueArray, numFields)

	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		bencodeKey(field, &svList[i])
		// The tag `bencode:"-"` should mean that this field must be ignored
		// See https://golang.org/pkg/encoding/json/#Marshal or https://golang.org/pkg/encoding/xml/#Marshal
		// We set a zero value so that it is ignored by the writeSVList() function
		if svList[i].key == "-" {
			svList[i].value = reflect.Value{}
		} else {
			svList[i].value = val.Field(i)
		}
	}

	err = writeSVList(w, svList)
	if err != nil {
		return
	}

	_, err = fmt.Fprint(w, "e")
	if err != nil {
		return
	}
	return
}

func writeValue(w io.Writer, val reflect.Value) (err error) {
	if !val.IsValid() {
		err = errors.New("can't write null value")
		return
	}

	switch v := val; v.Kind() {
	case reflect.String:
		s := v.String()
		_, err = fmt.Fprintf(w, "%d:%s", len(s), s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = fmt.Fprintf(w, "i%de", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		_, err = fmt.Fprintf(w, "i%de", v.Uint())
	case reflect.Array:
		err = writeArrayOrSlice(w, v)
	case reflect.Slice:
		switch val.Type().String() {
		case "[]uint8":
			// special case as byte-string
			s := string(v.Bytes())
			_, err = fmt.Fprintf(w, "%d:%s", len(s), s)
		default:
			err = writeArrayOrSlice(w, v)
		}
	case reflect.Map:
		err = writeMap(w, v)
	case reflect.Struct:
		err = writeStruct(w, v)
	case reflect.Interface:
		err = writeValue(w, v.Elem())
	default:
		err = &MarshalError{val.Type()}
	}
	return
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		fmt.Println("oasdas")
	}
	return false
}

func (sv stringValue) isValueNil() bool {
	if !sv.value.IsValid() || (sv.omitEmpty && isEmptyValue(sv.value)) {
		return true
	}
	switch v := sv.value; v.Kind() {
	case reflect.Interface:
		return !v.Elem().IsValid()
	}
	return false
}

// Marshal writes the bencode encoding of val to w.
//
// Marshal traverses the value v recursively.
//
// Marshal uses the following type-dependent encodings:
//
// Floating point, integer, and Number values encode as bencode numbers.
//
// String values encode as bencode strings.
//
// Array and slice values encode as bencode arrays.
//
// Struct values encode as bencode maps. Each exported struct field
// becomes a member of the object.
// The object's default key string is the struct field name
// but can be specified in the struct field's tag value. The text of
// the struct field's tag value is the key name. Examples:
//
//	// Field appears in bencode as key "Field".
//	Field int
//
//	// Field appears in bencode as key "myName".
//	Field int "myName"
//
// Anonymous struct fields are ignored.
//
// Map values encode as bencode objects.
// The map's key type must be string; the object keys are used directly
// as map keys.
//
// Boolean, Pointer, Interface, Channel, complex, and function values cannot
// be encoded in bencode.
// Attempting to encode such a value causes Marshal to return
// a MarshalError.
//
// Bencode cannot represent cyclic data structures and Marshal does not
// handle them.  Passing cyclic structures to Marshal will result in
// an infinite recursion.
func Marshal(w io.Writer, val interface{}) error {
	return writeValue(w, reflect.ValueOf(val))
}
