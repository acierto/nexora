// Package yamlutil provides strict, typed YAML decoding shared by configuration
// and Markdown metadata. It is deliberately separate from the domain model.
package yamlutil

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// Decode rejects unknown fields, duplicate keys, nulls and scalar coercion.
func Decode(data []byte, target any) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("invalid UTF-8")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("only one YAML document is allowed")
	}
	if len(document.Content) != 1 {
		return fmt.Errorf("expected one YAML document")
	}
	typ := reflect.TypeOf(target)
	if typ == nil || typ.Kind() != reflect.Pointer {
		return fmt.Errorf("YAML target must be a pointer")
	}
	if err := validate(document.Content[0], typ.Elem(), "", map[*yaml.Node]bool{}); err != nil {
		return err
	}
	return document.Decode(target)
}

func validate(n *yaml.Node, typ reflect.Type, field string, visiting map[*yaml.Node]bool) error {
	if visiting[n] {
		return fmt.Errorf("%s: cyclic YAML alias", field)
	}
	visiting[n] = true
	defer delete(visiting, n)
	if n.Kind == yaml.AliasNode {
		return validate(n.Alias, typ, field, visiting)
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	fail := func(message string) error { return fmt.Errorf("line %d, %s: %s", n.Line, field, message) }
	if n.Tag == "!!null" {
		return fail("null is not allowed; omit optional fields instead")
	}
	switch typ.Kind() {
	case reflect.Struct:
		if n.Kind != yaml.MappingNode || n.Tag != "!!map" {
			return fail("expected a mapping")
		}
		fields := map[string]reflect.Type{}
		for idx := 0; idx < typ.NumField(); idx++ {
			f := typ.Field(idx)
			tag := strings.Split(f.Tag.Get("yaml"), ",")[0]
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = strings.ToLower(f.Name)
			}
			fields[tag] = f.Type
		}
		seen := map[string]bool{}
		for idx := 0; idx < len(n.Content); idx += 2 {
			key, value := n.Content[idx], n.Content[idx+1]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return fail("mapping keys must be strings")
			}
			if seen[key.Value] {
				return fail("duplicate key " + key.Value)
			}
			seen[key.Value] = true
			ft, ok := fields[key.Value]
			if !ok {
				return fail("unknown field " + key.Value)
			}
			name := key.Value
			if field != "" {
				name = field + "." + name
			}
			if err := validate(value, ft, name, visiting); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if n.Kind != yaml.SequenceNode || n.Tag != "!!seq" {
			return fail("expected a sequence")
		}
		for idx, value := range n.Content {
			if err := validate(value, typ.Elem(), fmt.Sprintf("%s[%d]", field, idx), visiting); err != nil {
				return err
			}
		}
	case reflect.String:
		if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
			return fail("expected a string (quote numeric versions)")
		}
	case reflect.Int:
		if n.Kind != yaml.ScalarNode || n.Tag != "!!int" {
			return fail("expected an integer")
		}
	case reflect.Bool:
		if n.Kind != yaml.ScalarNode || n.Tag != "!!bool" {
			return fail("expected true or false")
		}
	default:
		return fail("unsupported YAML target type")
	}
	return nil
}
