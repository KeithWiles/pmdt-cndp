// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package raw

import (
	"bytes"
	"fmt"
	"reflect"
)

// Dump out a Go structure with offsets and length of the variables to allow
// matching offsets and sizes to a shared memory layout defined by a different
// language. Created to debug a C++ shared memory structure to Go.

// Debug function to help indent (spaces) in debug messages.
// Appends to a string for later dumping to console or file.
func (enc *EncodeRaw) indentStr() {
	enc.str += fmt.Sprintf("%s", bytes.Repeat([]byte("  "), enc.depth))
}

// Debug function to output indent spaces with a new line
// Appends to a string for later dumping to console or file.
func (enc *EncodeRaw) nlIndentStr() {
	enc.str += "\n"
	enc.indentStr()
}

// Debug function to output an open bracket for structures and arrays
// Appends to a string for later dumping to console or file.
func (enc *EncodeRaw) openBracketStr() {
	enc.str += " {\n"
	enc.depth++
	enc.indentStr()
}

// Debug function to output closing bracket.
// Appends to a string for later dumping to console or file.
func (enc *EncodeRaw) closeBracketStr() {
	enc.str += "}\n"
	enc.depth--
	enc.indentStr()
}

// Walk a reflect value and dump out information about the variable(s)
// Not all variable types in Go are supported.
func (enc *EncodeRaw) dumpWalk(v reflect.Value) error {

	kind := v.Kind()
	if kind == reflect.Invalid {
		return e("invalid type of value")
	}

	enc.str += fmt.Sprintf("%v", v.Type())

	switch v.Kind() {
	case reflect.Invalid:
		// Nothing to do here, move on

	// Output a newline and spaces for clearer information output
	case reflect.Bool, reflect.String, reflect.Interface,
		reflect.Float64, reflect.Float32, reflect.Complex64, reflect.Complex128,
		reflect.Uintptr, reflect.UnsafePointer, reflect.Chan, reflect.Func,
		reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
		reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		enc.nlIndentStr()

	// Walk all of the values in a slice or array call dumpWalk recursively
	case reflect.Slice, reflect.Array:
		enc.str += "\n"
		enc.indentStr()

		// Unpack the next level down variable as a reflect object
		nv := enc.unpackValue(v.Index(0))
		if nv.Kind() == reflect.Struct {
			if err := enc.dumpWalk(enc.unpackValue(nv)); err != nil {
				return err
			}
		}

	// A pointer type variable, which could be a structure or array or slice
	case reflect.Ptr:
		enc.nlIndentStr()
		if err := enc.dumpWalk(v.Elem()); err != nil {
			return err
		}

	// A MAP object with kery /value pairs
	case reflect.Map:
		if v.IsNil() {
			break
		}
		keys := v.MapKeys()
		enc.openBracketStr()
		for _, k := range keys {
			if err := enc.dumpWalk(enc.unpackValue(k)); err != nil {
				return err
			}
			if err := enc.dumpWalk(enc.unpackValue(v.MapIndex(k))); err != nil {
				return err
			}
		}
		enc.closeBracketStr()

	// A structure found walk all of the fields in the structure recursively
	case reflect.Struct:
		vt := v.Type()
		numFields := vt.NumField()
		enc.openBracketStr()
		for i := 0; i < numFields; i++ {
			vtf := vt.Field(i)
			if vtf.Name != "_" {
				enc.str += fmt.Sprintf("(%d).%s ", vtf.Offset, vtf.Name)
				if err := enc.dumpWalk(enc.unpackValue(v.Field(i))); err != nil {
					return err
				}
			}
		}
		enc.closeBracketStr()

	default:
		return fmt.Errorf("unknown type %v", v.Kind())
	}

	return nil
}

// Internal routine to start the dumping of the Go structure
// Pass an interface{} object to start the dump
// Return a string containing the dump text with offsets and sizes.
func (enc *EncodeRaw) dump(a ...interface{}) (string, error) {

	// Allow for a variable set of arguments to the dump command
	for _, arg := range a {
		rv := reflect.ValueOf(arg)

		if rv.Kind() != reflect.Ptr {
			continue
		}

		if rv.IsNil() {
			continue
		}

		// Process the variable recursively
		err := enc.dumpWalk(rv)
		if err != nil {
			return "", err
		}
	}
	enc.str += "\n"
	enc.depth = 0

	return enc.str, nil
}

// Dump for the given interface
// Global function to start the dumping of the objects.
func (enc *EncodeRaw) Dump(a ...interface{}) (string, error) {

	if a == nil {
		return "", e("dump value is nil")
	}

	return enc.dump(a...)
}
