// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package raw

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	tlog "pmdt.org/ttylog"
	"reflect"
	"regexp"
	"strconv"
)

// Encoding a shared memory segment into a Go structure.
//
// Not all types of variables are currently supported as the code was
// written for converting a simple C++ structure in shared memory.
//
// The code will encode meaning convert a byte array into a Go structure using
// Go's reflection package. Most likely this code will only work for the
// Performance monitor tool share memory layout.
//
// Written by Keith Wiles @ Intel Corp 2019

// For consistency:
var endian = binary.LittleEndian

// EncodeReader to handle grabbing the bytes
type EncodeReader interface {
	io.Reader
	io.ByteReader
}

// EncodeRaw information
type EncodeRaw struct {
	r         EncodeReader
	offset    uint64
	noPrint   bool
	depth     int
	len       int
	str       string
	fieldName string
}

var (
	// uint8Type is a reflect.Type representing a uint8.  It is used to
	// convert cgo types to uint8 slices for hexdumping.
	uint8Type = reflect.TypeOf(uint8(0))

	// cCharRE is a regular expression that matches a cgo char.
	// It is used to detect character arrays to hexdump them.
	cCharRE = regexp.MustCompile("^.*\\._Ctype_char$")

	// cUnsignedCharRE is a regular expression that matches a cgo unsigned
	// char.  It is used to detect unsigned character arrays to hexdump
	// them.
	cUnsignedCharRE = regexp.MustCompile("^.*\\._Ctype_unsignedchar$")

	// cUint8tCharRE is a regular expression that matches a cgo uint8_t.
	// It is used to detect uint8_t arrays to hexdump them.
	cUint8tCharRE = regexp.MustCompile("^.*\\._Ctype_uint8_t$")
)

// NewEncoder is the container for a raw encode of a structure
func NewEncoder(r EncodeReader) *EncodeRaw {

	enc := &EncodeRaw{r: r}

	return enc
}

func (enc *EncodeRaw) enable() {
	enc.noPrint = false
}

func (enc *EncodeRaw) disable() {
	enc.noPrint = true
}

// e a function helper to create error objects
func e(format string, args ...interface{}) error {
	return fmt.Errorf("encoding/raw: "+format, args...)
}

// unpackValue returns values inside of non-nil interfaces when possible.
// This is useful for data types like structs, arrays, slices, and maps which
// can contain varying types packed inside an interface.
func (enc *EncodeRaw) unpackValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// Encode a byte array to a Go Array structure
func (enc *EncodeRaw) encodeArray(v reflect.Value) error {

	for i := 0; i < v.Len(); i++ {
		nv := enc.unpackValue(v.Index(i))

		if err := enc.encode(nv); err != nil {
			tlog.DoPrintf("reflect.Array: %v\n", err)
			return err
		}
	}

	return nil
}

// Encode a byte array to a Go slice structure
func (enc *EncodeRaw) encodeSlice(v reflect.Value) error {

	size := v.Len()
	slice := reflect.MakeSlice(v.Type(), size, size)
	for i := 0; i < size; i++ {
		if err := enc.encode(slice.Index(i)); err != nil {
			return err
		}
	}
	v.Set(slice)

	return nil
}

// Encode a byte array to a Go int8 value
func (enc *EncodeRaw) int8Encoder(v reflect.Value) {
	var i8 int8

	// Binary encoding of the single byte as a Int8 value
	err := binary.Read(enc.r, binary.LittleEndian, &i8)
	if err != nil {
		tlog.ErrorPrintf("int8Encoder: error %v\n", err)
		return
	}
	v.SetInt(int64(i8))
	enc.offset++
}

// Encode a byte array to a Go int16 value
func (enc *EncodeRaw) int16Encoder(v reflect.Value) {
	var i16 int16

	// Binary encoding of the single 16 bit short as a Int16 value
	err := binary.Read(enc.r, binary.LittleEndian, &i16)
	if err != nil {
		tlog.ErrorPrintf("int16Encoder: error %v\n", err)
		return
	}
	v.SetInt(int64(i16))
	enc.offset += 2
}

// Encode a byte array to a Go int32 value
func (enc *EncodeRaw) int32Encoder(v reflect.Value) {
	var i32 int32

	// Binary encoding of the single 32 bit as a Int32 value
	err := binary.Read(enc.r, binary.LittleEndian, &i32)
	if err != nil {
		tlog.ErrorPrintf("int32Encoder: error %v\n", err)
		return
	}
	v.SetInt(int64(i32))
	enc.offset += 4
}

// Encode a byte array to a Go int64 value
func (enc *EncodeRaw) int64Encoder(v reflect.Value) {
	var i64 int64

	// Binary encoding of the single 64 bit as a Int64 value
	err := binary.Read(enc.r, binary.LittleEndian, &i64)
	if err != nil {
		tlog.ErrorPrintf("int64Encoder: error %v\n", err)
		return
	}
	v.SetInt(i64)
	enc.offset += 8
}

// Encode a byte array to a Go intptr value
func (enc *EncodeRaw) intptrEncoder(v reflect.Value) {
	if v.IsNil() {
		pzero := reflect.New(v.Type().Elem())
		v.Set(pzero)
	}
	// Call the encode again for the element type of the value
	enc.encode(v.Elem())
}

// Encode a byte array to a Go uint8 value
func (enc *EncodeRaw) uint8Encoder(v reflect.Value) {
	var u8 uint8

	// Binary encoding of the single byte as a Int8 value
	err := binary.Read(enc.r, binary.LittleEndian, &u8)
	if err != nil {
		tlog.ErrorPrintf("uint8Encoder: error %v\n", err)
		return
	}
	v.SetUint(uint64(u8))
	enc.offset++
}

// Encode a byte array to a Go uint16 value
func (enc *EncodeRaw) uint16Encoder(v reflect.Value) {
	var u16 uint16

	// Binary encoding of the single byte as a Int8 value
	err := binary.Read(enc.r, binary.LittleEndian, &u16)
	if err != nil {
		tlog.ErrorPrintf("uint16Encoder: error %v\n", err)
		return
	}
	v.SetUint(uint64(u16))
	enc.offset += 2
}

// Encode a byte array to a Go uint32 value
func (enc *EncodeRaw) uint32Encoder(v reflect.Value) {
	var u32 uint32

	// Binary encoding of the single 32 bit as a Uint32 value
	err := binary.Read(enc.r, binary.LittleEndian, &u32)
	if err != nil {
		tlog.ErrorPrintf("uint32Encoder: error %v\n", err)
		return
	}
	v.SetUint(uint64(u32))
	enc.offset += 4
}

// Encode a byte array to a Go uint64 value
func (enc *EncodeRaw) uint64Encoder(v reflect.Value) {
	var u64 uint64

	// Binary encoding of the single 64 bit as a Uint64 value
	err := binary.Read(enc.r, binary.LittleEndian, &u64)
	if err != nil {
		tlog.ErrorPrintf("uint64Encoder: error %v\n", err)
		return
	}
	v.SetUint(u64)
	enc.offset += 8
}

// Encode a byte array to a Go uintptr value
func (enc *EncodeRaw) uintptrEncoder(v reflect.Value) {
	if v.IsNil() {
		pzero := reflect.New(v.Type().Elem())
		v.Set(pzero)
	}
	// Call the encode again for the element type of the value
	enc.encode(v.Elem())
}

// Encode a byte array to a Go float32 value
func (enc *EncodeRaw) float32Encoder(v reflect.Value) {
	var u32 uint32

	// Binary encoding of the single 32 or 64 bit float as a Float32/64 value
	err := binary.Read(enc.r, binary.LittleEndian, &u32)
	if err != nil {
		tlog.ErrorPrintf("float32Encoder: error %v\n", err)
		return
	}
	v.SetFloat(math.Float64frombits(uint64(u32)))
	enc.offset += 4
}

// Encode a byte array to a Go float64 value
func (enc *EncodeRaw) float64Encoder(v reflect.Value) {
	var u64 uint64

	// Binary encoding of the single 32 or 64 bit float as a Float32/64 value
	err := binary.Read(enc.r, binary.LittleEndian, &u64)
	if err != nil {
		tlog.ErrorPrintf("float64Encoder: error %v\n", err)
		return
	}

	f64 := math.Float64frombits(u64)
	v.SetFloat(f64)
	enc.offset += 8
}

// Encode a byte array to a Go complex64 value
func (enc *EncodeRaw) complex64Encoder(v reflect.Value) {
	var u32rpart, u32ipart uint32

	// Binary encoding of the single 64 or 128 bit Complex as a Complex value
	err := binary.Read(enc.r, binary.LittleEndian, &u32rpart)
	if err != nil {
		tlog.ErrorPrintf("complex64Encoder: error %v\n", err)
		return
	}
	rpart := math.Float64frombits(uint64(u32rpart))
	err = binary.Read(enc.r, binary.LittleEndian, &u32ipart)
	if err != nil {
		tlog.ErrorPrintf("complex64Encoder: error %v\n", err)
		return
	}
	ipart := math.Float64frombits(uint64(u32ipart))
	v.SetComplex(complex(rpart, ipart))
	enc.offset += 8
}

// Encode a byte array to a Go complex128 value
func (enc *EncodeRaw) complex128Encoder(v reflect.Value) {
	var u64rpart, u64ipart uint64

	// Binary encoding of the single 64 or 128 bit Complex as a Complex value
	err := binary.Read(enc.r, binary.LittleEndian, &u64rpart)
	if err != nil {
		tlog.ErrorPrintf("complex128Encoder: error %v\n", err)
		return
	}
	rpart := math.Float64frombits(uint64(u64rpart))
	err = binary.Read(enc.r, binary.LittleEndian, &u64ipart)
	if err != nil {
		tlog.ErrorPrintf("complex128Encoder: error %v\n", err)
		return
	}
	ipart := math.Float64frombits(uint64(u64ipart))

	v.SetComplex(complex(rpart, ipart))
	enc.offset += 16
}

// Encode a byte array to a Go string value
func (enc *EncodeRaw) stringEncoder(v reflect.Value) {
	sz := v.Len()
	if sz == 0 {
		sz = enc.len
		enc.len = 0
	}
	// Create a local byte array for the size of the string
	b := make([]byte, sz)
	n, _ := enc.r.Read(b)

	// Convert the byte array to a slice, convert to string and store as a string
	v.SetString(string(b[:n]))
	enc.offset += uint64(n)
}

// Encode a byte array to a Go map value
func (enc *EncodeRaw) mapEncoder(v reflect.Value) {
	var count uint

	// Map processing and encoding the key and value
	if err := enc.encode(reflect.ValueOf(&count).Elem()); err != nil {
		tlog.ErrorPrintf("mapEncoder: error %v\n", err)
		return
	}

	// Create the map and set the key and value types
	mtyp := v.Type()
	ktyp := mtyp.Key()
	etyp := mtyp.Elem()
	v.Set(reflect.MakeMap(v.Type()))

	// Loop on the number of items in map converting to key/value pairs.
	for i := 0; i < int(count); i++ {
		key := reflect.New(ktyp).Elem()
		elem := reflect.New(etyp).Elem()

		// Determine type and encode the key value
		if err := enc.encode(key); err != nil {
			tlog.ErrorPrintf("mapEncoder: error %v\n", err)
			return
		}

		// Determine type and encode the value
		if err := enc.encode(elem); err != nil {
			tlog.ErrorPrintf("mapEncoder: error %v\n", err)
			return
		}

		// Store the element in the map using key value
		v.SetMapIndex(key, elem)
	}
}

// Encode a byte array to a Go struct value
func (enc *EncodeRaw) structEncoder(v reflect.Value) {
	styp := v.Type()

	// Process all of the fields in a structure skipping a Chan type.
	numFields := styp.NumField()
	for i := 0; i < numFields; i++ {
		field := styp.Field(i)
		if field.Type.Kind() == reflect.Chan {
			continue
		}
		enc.fieldName = field.Name
		enc.len, _ = strconv.Atoi(field.Tag.Get("rawlen"))
		if err := enc.encode(v.Field(i)); err != nil {
			tlog.ErrorPrintf("structEncoder: error %v\n", err)
			return
		}
	}
}

// Encode a byte array to a Go bool value
func (enc *EncodeRaw) boolEncoder(v reflect.Value) {
	c, err := enc.r.ReadByte()
	if err != nil {
		tlog.ErrorPrintf("boolEncoder: error %v\n", err)
		return
	}
	v.SetBool(c == 1)
	enc.offset++
}

// encode a byte array in a Go structure
func (enc *EncodeRaw) encode(v reflect.Value) error {

	kind := v.Kind()

	switch kind {
	case reflect.Invalid:
		// Nothing to do here, move on

	case reflect.Ptr:
		enc.intptrEncoder(v)

	case reflect.Int8:
		enc.int8Encoder(v)

	case reflect.Int16:
		enc.int16Encoder(v)

	case reflect.Int32:
		enc.int32Encoder(v)

	case reflect.Int, reflect.Int64:
		enc.int64Encoder(v)

	case reflect.Uint8:
		enc.uint8Encoder(v)

	case reflect.Uint16:
		enc.uint16Encoder(v)

	case reflect.Uint32:
		enc.uint32Encoder(v)

	case reflect.Uint, reflect.Uint64:
		enc.uint64Encoder(v)

	case reflect.Uintptr:
		enc.uintptrEncoder(v)

	case reflect.Float32:
		enc.float32Encoder(v)

	case reflect.Float64:
		enc.float64Encoder(v)

	case reflect.Complex64:
		enc.complex64Encoder(v)

	case reflect.Complex128:
		enc.complex128Encoder(v)

	case reflect.Array:
		enc.encodeArray(v)

	case reflect.String:
		enc.stringEncoder(v)

	case reflect.Slice:
		enc.encodeSlice(v)

	case reflect.Map:
		enc.mapEncoder(v)

	case reflect.Struct:
		enc.structEncoder(v)

	case reflect.Bool:
		enc.boolEncoder(v)

	default:
		tlog.ErrorPrintf("Unknown type %v\n", v.Kind())
	}

	return nil
}

// Encode the shared memory region into SharedPCMState structure
func (enc *EncodeRaw) Encode(v interface{}) error {

	if v == nil {
		return e("encode to value is nil")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return e("value is not a pointer")
	}

	if rv.IsNil() {
		return e("pointer is nil")
	}

	enc.disable()
	enc.fieldName = "SharedPCMState"
	if err := enc.encode(rv); err != nil {
		return err
	}
	enc.enable()

	return nil
}
