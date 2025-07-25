// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

type marshalTest struct {
	cborData []byte
	values   []any
}

type marshalErrorTest struct {
	name         string
	value        any
	wantErrorMsg string
}

type inner struct {
	X, Y, z int64
}

type outer struct {
	IntField          int
	FloatField        float32
	BoolField         bool
	StringField       string
	ByteStringField   []byte
	ArrayField        []string
	MapField          map[string]bool
	NestedStructField *inner
	unexportedField   int64
}

// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
var marshalTests = []marshalTest{
	// positive integer
	{hexDecode("00"), []any{uint(0), uint8(0), uint16(0), uint32(0), uint64(0), int(0), int8(0), int16(0), int32(0), int64(0)}},
	{hexDecode("01"), []any{uint(1), uint8(1), uint16(1), uint32(1), uint64(1), int(1), int8(1), int16(1), int32(1), int64(1)}},
	{hexDecode("0a"), []any{uint(10), uint8(10), uint16(10), uint32(10), uint64(10), int(10), int8(10), int16(10), int32(10), int64(10)}},
	{hexDecode("17"), []any{uint(23), uint8(23), uint16(23), uint32(23), uint64(23), int(23), int8(23), int16(23), int32(23), int64(23)}},
	{hexDecode("1818"), []any{uint(24), uint8(24), uint16(24), uint32(24), uint64(24), int(24), int8(24), int16(24), int32(24), int64(24)}},
	{hexDecode("1819"), []any{uint(25), uint8(25), uint16(25), uint32(25), uint64(25), int(25), int8(25), int16(25), int32(25), int64(25)}},
	{hexDecode("1864"), []any{uint(100), uint8(100), uint16(100), uint32(100), uint64(100), int(100), int8(100), int16(100), int32(100), int64(100)}},
	{hexDecode("18ff"), []any{uint(255), uint8(255), uint16(255), uint32(255), uint64(255), int(255), int16(255), int32(255), int64(255)}},
	{hexDecode("190100"), []any{uint(256), uint16(256), uint32(256), uint64(256), int(256), int16(256), int32(256), int64(256)}},
	{hexDecode("1903e8"), []any{uint(1000), uint16(1000), uint32(1000), uint64(1000), int(1000), int16(1000), int32(1000), int64(1000)}},
	{hexDecode("19ffff"), []any{uint(65535), uint16(65535), uint32(65535), uint64(65535), int(65535), int32(65535), int64(65535)}},
	{hexDecode("1a00010000"), []any{uint(65536), uint32(65536), uint64(65536), int(65536), int32(65536), int64(65536)}},
	{hexDecode("1a000f4240"), []any{uint(1000000), uint32(1000000), uint64(1000000), int(1000000), int32(1000000), int64(1000000)}},
	{hexDecode("1affffffff"), []any{uint(4294967295), uint32(4294967295), uint64(4294967295), int64(4294967295)}},
	{hexDecode("1b000000e8d4a51000"), []any{uint64(1000000000000), int64(1000000000000)}},
	{hexDecode("1bffffffffffffffff"), []any{uint64(18446744073709551615)}},
	// negative integer
	{hexDecode("20"), []any{int(-1), int8(-1), int16(-1), int32(-1), int64(-1)}},
	{hexDecode("29"), []any{int(-10), int8(-10), int16(-10), int32(-10), int64(-10)}},
	{hexDecode("37"), []any{int(-24), int8(-24), int16(-24), int32(-24), int64(-24)}},
	{hexDecode("3818"), []any{int(-25), int8(-25), int16(-25), int32(-25), int64(-25)}},
	{hexDecode("3863"), []any{int(-100), int8(-100), int16(-100), int32(-100), int64(-100)}},
	{hexDecode("38ff"), []any{int(-256), int16(-256), int32(-256), int64(-256)}},
	{hexDecode("390100"), []any{int(-257), int16(-257), int32(-257), int64(-257)}},
	{hexDecode("3903e7"), []any{int(-1000), int16(-1000), int32(-1000), int64(-1000)}},
	{hexDecode("39ffff"), []any{int(-65536), int32(-65536), int64(-65536)}},
	{hexDecode("3a00010000"), []any{int(-65537), int32(-65537), int64(-65537)}},
	{hexDecode("3affffffff"), []any{int64(-4294967296)}},
	// byte string
	{hexDecode("40"), []any{[]byte{}}},
	{hexDecode("4401020304"), []any{[]byte{1, 2, 3, 4}, [...]byte{1, 2, 3, 4}}},
	// text string
	{hexDecode("60"), []any{""}},
	{hexDecode("6161"), []any{"a"}},
	{hexDecode("6449455446"), []any{"IETF"}},
	{hexDecode("62225c"), []any{"\"\\"}},
	{hexDecode("62c3bc"), []any{"ü"}},
	{hexDecode("63e6b0b4"), []any{"水"}},
	{hexDecode("64f0908591"), []any{"𐅑"}},
	// array
	{
		hexDecode("80"),
		[]any{
			[0]int{},
			[]uint{},
			// []uint8{},
			[]uint16{},
			[]uint32{},
			[]uint64{},
			[]int{},
			[]int8{},
			[]int16{},
			[]int32{},
			[]int64{},
			[]string{},
			[]bool{}, []float32{}, []float64{}, []any{},
		},
	},
	{
		hexDecode("83010203"),
		[]any{
			[...]int{1, 2, 3},
			[]uint{1, 2, 3},
			// []uint8{1, 2, 3},
			[]uint16{1, 2, 3},
			[]uint32{1, 2, 3},
			[]uint64{1, 2, 3},
			[]int{1, 2, 3},
			[]int8{1, 2, 3},
			[]int16{1, 2, 3},
			[]int32{1, 2, 3},
			[]int64{1, 2, 3},
			[]any{1, 2, 3},
		},
	},
	{
		hexDecode("8301820203820405"),
		[]any{
			[...]any{1, [...]int{2, 3}, [...]int{4, 5}},
			[]any{1, []uint{2, 3}, []uint{4, 5}},
			// []any{1, []uint8{2, 3}, []uint8{4, 5}},
			[]any{1, []uint16{2, 3}, []uint16{4, 5}},
			[]any{1, []uint32{2, 3}, []uint32{4, 5}},
			[]any{1, []uint64{2, 3}, []uint64{4, 5}},
			[]any{1, []int{2, 3}, []int{4, 5}},
			[]any{1, []int8{2, 3}, []int8{4, 5}},
			[]any{1, []int16{2, 3}, []int16{4, 5}},
			[]any{1, []int32{2, 3}, []int32{4, 5}},
			[]any{1, []int64{2, 3}, []int64{4, 5}},
			[]any{1, []any{2, 3}, []any{4, 5}},
		},
	},
	{
		hexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
		[]any{
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			// []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
		},
	},
	{
		hexDecode("826161a161626163"),
		[]any{
			[...]any{"a", map[string]string{"b": "c"}},
			[]any{"a", map[string]string{"b": "c"}},
			[]any{"a", map[any]any{"b": "c"}},
		},
	},
	// map
	{
		hexDecode("a0"),
		[]any{
			map[uint]bool{},
			map[uint8]bool{},
			map[uint16]bool{},
			map[uint32]bool{},
			map[uint64]bool{},
			map[int]bool{},
			map[int8]bool{},
			map[int16]bool{},
			map[int32]bool{},
			map[int64]bool{},
			map[float32]bool{},
			map[float64]bool{},
			map[bool]bool{},
			map[string]bool{},
			map[any]any{},
		},
	},
	{
		hexDecode("a201020304"),
		[]any{
			map[uint]uint{3: 4, 1: 2},
			map[uint8]uint8{3: 4, 1: 2},
			map[uint16]uint16{3: 4, 1: 2},
			map[uint32]uint32{3: 4, 1: 2},
			map[uint64]uint64{3: 4, 1: 2},
			map[int]int{3: 4, 1: 2},
			map[int8]int8{3: 4, 1: 2},
			map[int16]int16{3: 4, 1: 2},
			map[int32]int32{3: 4, 1: 2},
			map[int64]int64{3: 4, 1: 2},
			map[any]any{3: 4, 1: 2},
		},
	},
	{
		hexDecode("a26161016162820203"),
		[]any{
			map[string]any{"a": 1, "b": []any{2, 3}},
			map[any]any{"b": []any{2, 3}, "a": 1},
		},
	},
	{
		hexDecode("a56161614161626142616361436164614461656145"),
		[]any{
			map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[any]any{"b": "B", "a": "A", "c": "C", "e": "E", "d": "D"},
		},
	},
	// tag
	{
		hexDecode("c074323031332d30332d32315432303a30343a30305a"),
		[]any{Tag{0, "2013-03-21T20:04:00Z"}, RawTag{0, hexDecode("74323031332d30332d32315432303a30343a30305a")}},
	}, // 0: standard date/time
	{
		hexDecode("c11a514b67b0"),
		[]any{Tag{1, uint64(1363896240)}, RawTag{1, hexDecode("1a514b67b0")}},
	}, // 1: epoch-based date/time
	{
		hexDecode("c249010000000000000000"),
		[]any{
			bigIntOrPanic("18446744073709551616"),
			Tag{2, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{2, hexDecode("49010000000000000000")},
		},
	}, // 2: positive bignum: 18446744073709551616
	{
		hexDecode("c349010000000000000000"),
		[]any{
			bigIntOrPanic("-18446744073709551617"),
			Tag{3, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{3, hexDecode("49010000000000000000")},
		},
	}, // 3: negative bignum: -18446744073709551617
	{
		hexDecode("c1fb41d452d9ec200000"),
		[]any{Tag{1, float64(1363896240.5)}, RawTag{1, hexDecode("fb41d452d9ec200000")}},
	}, // 1: epoch-based date/time
	{
		hexDecode("d74401020304"),
		[]any{Tag{23, []byte{0x01, 0x02, 0x03, 0x04}}, RawTag{23, hexDecode("4401020304")}},
	}, // 23: expected conversion to base16 encoding
	{
		hexDecode("d818456449455446"),
		[]any{Tag{24, []byte{0x64, 0x49, 0x45, 0x54, 0x46}}, RawTag{24, hexDecode("456449455446")}},
	}, // 24: encoded cborBytes data item
	{
		hexDecode("d82076687474703a2f2f7777772e6578616d706c652e636f6d"),
		[]any{Tag{32, "http://www.example.com"}, RawTag{32, hexDecode("76687474703a2f2f7777772e6578616d706c652e636f6d")}},
	}, // 32: URI
	// primitives
	{hexDecode("f4"), []any{false}},
	{hexDecode("f5"), []any{true}},
	{hexDecode("f6"), []any{nil, []byte(nil), []int(nil), map[uint]bool(nil), (*int)(nil), io.Reader(nil)}},
	// nan, positive and negative inf
	{hexDecode("f97c00"), []any{math.Inf(1)}},
	{hexDecode("f97e00"), []any{math.NaN()}},
	{hexDecode("f9fc00"), []any{math.Inf(-1)}},
	// float32
	{hexDecode("fa47c35000"), []any{float32(100000.0)}},
	{hexDecode("fa7f7fffff"), []any{float32(3.4028234663852886e+38)}},
	// float64
	{hexDecode("fb3ff199999999999a"), []any{float64(1.1)}},
	{hexDecode("fb7e37e43c8800759c"), []any{float64(1.0e+300)}},
	{hexDecode("fbc010666666666666"), []any{float64(-4.1)}},
	// More testcases not covered by https://tools.ietf.org/html/rfc7049#appendix-A.
	{
		hexDecode("d83dd183010203"), // 61(17([1, 2, 3])), nested tags 61 and 17
		[]any{Tag{61, Tag{17, []any{uint64(1), uint64(2), uint64(3)}}}, RawTag{61, hexDecode("d183010203")}},
	},
}

var exMarshalTests = []marshalTest{
	{
		// array of nils
		hexDecode("83f6f6f6"),
		[]any{
			[]any{nil, nil, nil},
		},
	},
}

func TestMarshal(t *testing.T) {
	testMarshal(t, marshalTests)
	testMarshal(t, exMarshalTests)
}

func TestInvalidTypeMarshal(t *testing.T) {
	type s1 struct {
		Chan chan bool
	}
	type s2 struct {
		_    struct{} `cbor:",toarray"`
		Chan chan bool
	}
	var marshalErrorTests = []marshalErrorTest{
		{"channel cannot be marshaled", make(chan bool), "cbor: unsupported type: chan bool"},
		{"slice of channel cannot be marshaled", make([]chan bool, 10), "cbor: unsupported type: []chan bool"},
		{"slice of pointer to channel cannot be marshaled", make([]*chan bool, 10), "cbor: unsupported type: []*chan bool"},
		{"map of channel cannot be marshaled", make(map[string]chan bool), "cbor: unsupported type: map[string]chan bool"},
		{"struct of channel cannot be marshaled", s1{}, "cbor: unsupported type: cbor.s1"},
		{"struct of channel cannot be marshaled", s2{}, "cbor: unsupported type: cbor.s2"},
		{"function cannot be marshaled", func(i int) int { return i * i }, "cbor: unsupported type: func(int) int"},
		{"complex cannot be marshaled", complex(100, 8), "cbor: unsupported type: complex128"},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range marshalErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(&tc.value)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*UnsupportedTypeError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *UnsupportedTypeError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			} else if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}

			b, err = em.Marshal(&tc.value)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*UnsupportedTypeError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *UnsupportedTypeError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			} else if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}
		})
	}
}

func TestMarshalLargeByteString(t *testing.T) {
	// []byte{100, 100, 100, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 10000000}
	tests := make([]marshalTest, len(lengths))
	for i, length := range lengths {
		cborData := bytes.NewBuffer(encodeCborHeader(cborTypeByteString, uint64(length)))
		value := make([]byte, length)
		for j := 0; j < length; j++ {
			cborData.WriteByte(100)
			value[j] = 100
		}
		tests[i] = marshalTest{cborData.Bytes(), []any{value}}
	}

	testMarshal(t, tests)
}

func TestMarshalLargeTextString(t *testing.T) {
	// "ddd..."
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 10000000}
	tests := make([]marshalTest, len(lengths))
	for i, length := range lengths {
		cborData := bytes.NewBuffer(encodeCborHeader(cborTypeTextString, uint64(length)))
		value := make([]byte, length)
		for j := 0; j < length; j++ {
			cborData.WriteByte(100)
			value[j] = 100
		}
		tests[i] = marshalTest{cborData.Bytes(), []any{string(value)}}
	}

	testMarshal(t, tests)
}

func TestMarshalLargeArray(t *testing.T) {
	// []string{"水", "水", "水", ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	tests := make([]marshalTest, len(lengths))
	for i, length := range lengths {
		cborData := bytes.NewBuffer(encodeCborHeader(cborTypeArray, uint64(length)))
		value := make([]string, length)
		for j := 0; j < length; j++ {
			cborData.Write([]byte{0x63, 0xe6, 0xb0, 0xb4})
			value[j] = "水"
		}
		tests[i] = marshalTest{cborData.Bytes(), []any{value}}
	}

	testMarshal(t, tests)
}

func TestMarshalLargeMapCanonical(t *testing.T) {
	// map[int]int {0:0, 1:1, 2:2, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	tests := make([]marshalTest, len(lengths))
	for i, length := range lengths {
		cborData := bytes.NewBuffer(encodeCborHeader(cborTypeMap, uint64(length)))
		value := make(map[int]int, length)
		for j := 0; j < length; j++ {
			d := encodeCborHeader(cborTypePositiveInt, uint64(j))
			cborData.Write(d)
			cborData.Write(d)
			value[j] = j
		}
		tests[i] = marshalTest{cborData.Bytes(), []any{value}}
	}

	testMarshal(t, tests)
}

func TestMarshalLargeMap(t *testing.T) {
	// map[int]int {0:0, 1:1, 2:2, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	for _, length := range lengths {
		m1 := make(map[int]int, length)
		for i := 0; i < length; i++ {
			m1[i] = i
		}

		cborData, err := Marshal(m1)
		if err != nil {
			t.Fatalf("Marshal(%v) returned error %v", m1, err)
		}

		m2 := make(map[int]int)
		if err = Unmarshal(cborData, &m2); err != nil {
			t.Fatalf("Unmarshal(0x%x) returned error %v", cborData, err)
		}

		if !reflect.DeepEqual(m1, m2) {
			t.Errorf("Unmarshal() = %v, want %v", m2, m1)
		}
	}
}

func encodeCborHeader(t cborType, n uint64) []byte {
	b := make([]byte, 9)
	switch {
	case n <= 23:
		b[0] = byte(t) | byte(n)
		return b[:1]
	case n <= math.MaxUint8:
		b[0] = byte(t) | byte(24)
		b[1] = byte(n)
		return b[:2]
	case n <= math.MaxUint16:
		b[0] = byte(t) | byte(25)
		binary.BigEndian.PutUint16(b[1:], uint16(n))
		return b[:3]
	case n <= math.MaxUint32:
		b[0] = byte(t) | byte(26)
		binary.BigEndian.PutUint32(b[1:], uint32(n))
		return b[:5]
	default:
		b[0] = byte(t) | byte(27)
		binary.BigEndian.PutUint64(b[1:], n)
		return b[:9]
	}
}

func testMarshal(t *testing.T, testCases []marshalTest) {
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		for _, value := range tc.values {
			if _, err := Marshal(value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", value, err)
			}
			if b, err := em.Marshal(value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", value, err)
			} else if !bytes.Equal(b, tc.cborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", value, b, tc.cborData)
			}
		}
		r := RawMessage(tc.cborData)
		if b, err := Marshal(r); err != nil {
			t.Errorf("Marshal(%v) returned error %v", r, err)
		} else if !bytes.Equal(b, r) {
			t.Errorf("Marshal(%v) returned %v, want %v", r, b, r)
		}
	}
}

func TestMarshalStruct(t *testing.T) {
	v1 := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000, z: 10000000},
		unexportedField:   6,
	}
	unmarshalWant := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000},
	}

	cborData, err := Marshal(v1)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v1, err)
	}

	var v2 outer
	if err = Unmarshal(cborData, &v2); err != nil {
		t.Fatalf("Unmarshal(0x%x) returned error %v", cborData, err)
	}

	if !reflect.DeepEqual(unmarshalWant, v2) {
		t.Errorf("Unmarshal() = %v, want %v", v2, unmarshalWant)
	}
}
func TestMarshalStructCanonical(t *testing.T) {
	v := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000, z: 10000000},
		unexportedField:   6,
	}
	var cborData bytes.Buffer
	cborData.WriteByte(byte(cborTypeMap) | 8) // CBOR header: map type with 8 items (exported fields)

	cborData.WriteByte(byte(cborTypeTextString) | 8) // "IntField"
	cborData.WriteString("IntField")
	cborData.WriteByte(byte(cborTypePositiveInt) | 24)
	cborData.WriteByte(123)

	cborData.WriteByte(byte(cborTypeTextString) | 8) // "MapField"
	cborData.WriteString("MapField")
	cborData.WriteByte(byte(cborTypeMap) | 2)
	cborData.WriteByte(byte(cborTypeTextString) | 7)
	cborData.WriteString("morning")
	cborData.WriteByte(byte(cborTypePrimitives) | 21)
	cborData.WriteByte(byte(cborTypeTextString) | 9)
	cborData.WriteString("afternoon")
	cborData.WriteByte(byte(cborTypePrimitives) | 20)

	cborData.WriteByte(byte(cborTypeTextString) | 9) // "BoolField"
	cborData.WriteString("BoolField")
	cborData.WriteByte(byte(cborTypePrimitives) | 21)

	cborData.WriteByte(byte(cborTypeTextString) | 10) // "ArrayField"
	cborData.WriteString("ArrayField")
	cborData.WriteByte(byte(cborTypeArray) | 2)
	cborData.WriteByte(byte(cborTypeTextString) | 5)
	cborData.WriteString("hello")
	cborData.WriteByte(byte(cborTypeTextString) | 5)
	cborData.WriteString("world")

	cborData.WriteByte(byte(cborTypeTextString) | 10) // "FloatField"
	cborData.WriteString("FloatField")
	cborData.Write([]byte{0xfa, 0x47, 0xc3, 0x50, 0x00})

	cborData.WriteByte(byte(cborTypeTextString) | 11) // "StringField"
	cborData.WriteString("StringField")
	cborData.WriteByte(byte(cborTypeTextString) | 4)
	cborData.WriteString("test")

	cborData.WriteByte(byte(cborTypeTextString) | 15) // "ByteStringField"
	cborData.WriteString("ByteStringField")
	cborData.WriteByte(byte(cborTypeByteString) | 3)
	cborData.Write([]byte{1, 3, 5})

	cborData.WriteByte(byte(cborTypeTextString) | 17) // "NestedStructField"
	cborData.WriteString("NestedStructField")
	cborData.WriteByte(byte(cborTypeMap) | 2)
	cborData.WriteByte(byte(cborTypeTextString) | 1)
	cborData.WriteString("X")
	cborData.WriteByte(byte(cborTypePositiveInt) | 25)
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(1000))
	cborData.Write(b)
	cborData.WriteByte(byte(cborTypeTextString) | 1)
	cborData.WriteString("Y")
	cborData.WriteByte(byte(cborTypePositiveInt) | 26)
	b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(1000000))
	cborData.Write(b)

	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %q", err)
	}
	if b, err := em.Marshal(v); err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, cborData.Bytes()) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, cborData.Bytes())
	}
}

func TestMarshalNullPointerToEmbeddedStruct(t *testing.T) {
	type (
		T1 struct {
			X int
		}
		T2 struct {
			*T1
		}
	)
	v := T2{}
	wantCborData := []byte{0xa0} // {}
	cborData, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, cborData) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, cborData, wantCborData)
	}
}

func TestMarshalNullPointerToStruct(t *testing.T) {
	type (
		T1 struct {
			X int
		}
		T2 struct {
			T *T1
		}
	)
	v := T2{}
	wantCborData := []byte{0xa1, 0x61, 0x54, 0xf6} // {X: nil}
	cborData, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, cborData) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, cborData, wantCborData)
	}
}

// Struct fields encoding follows the same struct fields visibility
// rules used by JSON encoding package.  Some struct types are from
// tests in JSON encoding package to ensure that the same rules are
// followed.
func TestAnonymousFields1(t *testing.T) {
	// Fields (T1.X, T2.X) with the same name at the same level are ignored
	type (
		T1 struct{ x, X int }
		T2 struct{ x, X int }
		T  struct {
			T1
			T2
		}
	)
	v := T{T1{1, 2}, T2{3, 4}}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields2(t *testing.T) {
	// Field (T.X) with the same name at a less nested level is serialized
	type (
		T1 struct{ x, X int }
		T2 struct{ x, X int }
		T  struct {
			T1
			T2
			x, X int
		}
	)
	v := T{T1{1, 2}, T2{3, 4}, 5, 6}
	want := []byte{0xa1, 0x61, 0x58, 0x06} // {X:6}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{X: 6}
	if err := Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields3(t *testing.T) {
	// Unexported embedded field (myInt) of non-struct type is ignored
	type (
		myInt int
		T     struct {
			myInt
		}
	)
	v := T{5}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields4(t *testing.T) {
	// Exported embedded field (MyInt) of non-struct type is serialized
	type (
		MyInt int
		T     struct {
			MyInt
		}
	)
	v := T{5}
	want := []byte{0xa1, 0x65, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x05} // {MyInt: 5}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v, v2) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v, v, v2, v2)
	}
}

func TestAnonymousFields5(t *testing.T) {
	// Unexported embedded field (*myInt) of pointer to non-struct type is ignored
	type (
		myInt int
		T     struct {
			*myInt
		}
	)
	v := T{new(myInt)}
	*v.myInt = 5
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields6(t *testing.T) {
	// Exported embedded field (*MyInt) of pointer to non-struct type should be serialized
	type (
		MyInt int
		T     struct {
			*MyInt
		}
	)
	v := T{new(MyInt)}
	*v.MyInt = 5
	want := []byte{0xa1, 0x65, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x05} // {MyInt: 5}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v, v2) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v, v, v2, v2)
	}
}

func TestAnonymousFields7(t *testing.T) {
	// Exported fields (t1.X, T2.Y) of embedded structs should have their exported fields be serialized
	type (
		t1 struct{ x, X int }
		T2 struct{ y, Y int }
		T  struct {
			t1
			T2
		}
	)
	v := T{t1{1, 2}, T2{3, 4}}
	want := []byte{0xa2, 0x61, 0x58, 0x02, 0x61, 0x59, 0x04} // {X:2, Y:4}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1{X: 2}, T2{Y: 4}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields8(t *testing.T) {
	// Exported fields of pointers (t1.X, T2.Y)
	type (
		t1 struct{ x, X int }
		T2 struct{ y, Y int }
		T  struct {
			*t1
			*T2
		}
	)
	v := T{&t1{1, 2}, &T2{3, 4}}
	want := []byte{0xa2, 0x61, 0x58, 0x02, 0x61, 0x59, 0x04} // {X:2, Y:4}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	// v1 cannot be unmarshaled to because reflect cannot allocate unexported field s1.
	var v1 T
	wantErrorMsg := "cannot set embedded pointer to unexported struct"
	wantV := T{T2: &T2{Y: 4}}
	err = Unmarshal(b, &v1)
	if err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error %q", b, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error %q", b, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v1, wantV) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", b, v1, v1, wantV, wantV)
	}

	// v2 can be unmarshaled to because unexported field t1 is already allocated.
	var v2 T
	v2.t1 = &t1{}
	unmarshalWant := T{&t1{X: 2}, &T2{Y: 4}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields9(t *testing.T) {
	// Multiple levels of nested anonymous fields
	type (
		MyInt1 int
		MyInt2 int
		myInt  int
		t2     struct {
			MyInt2
			myInt
		}
		t1 struct {
			MyInt1
			myInt
			t2
		}
		T struct {
			t1
			myInt
		}
	)
	v := T{t1{1, 2, t2{3, 4}}, 6}
	want := []byte{0xa2, 0x66, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x31, 0x01, 0x66, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x32, 0x03} // {MyInt1: 1, MyInt2: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{MyInt1: 1, t2: t2{MyInt2: 3}}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields10(t *testing.T) {
	// Fields of the same struct type at the same level
	type (
		t3 struct {
			Z int
		}
		t1 struct {
			X int
			t3
		}
		t2 struct {
			Y int
			t3
		}
		T struct {
			t1
			t2
		}
	)
	v := T{t1{1, t3{2}}, t2{3, t3{4}}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x61, 0x59, 0x03} // {X: 1, Y: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{X: 1}, t2: t2{Y: 3}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields11(t *testing.T) {
	// Fields (T.t2.X, T.t1.t2.X) of the same struct type at different levels
	type (
		t2 struct {
			X int
		}
		t1 struct {
			Y int
			t2
		}
		T struct {
			t1
			t2
		}
	)
	v := T{t1{1, t2{2}}, t2{3}}
	want := []byte{0xa2, 0x61, 0x59, 0x01, 0x61, 0x58, 0x03} // {Y: 1, X: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{Y: 1}, t2: t2{X: 3}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestOmitAndRenameStructField(t *testing.T) {
	type T struct {
		I   int // never omit
		Io  int `cbor:",omitempty"` // omit empty
		Iao int `cbor:"-"`          // always omit
		R   int `cbor:"omitempty"`  // renamed to omitempty
	}

	v1 := T{}
	// {"I": 0, "omitempty": 0}
	want1 := []byte{0xa2,
		0x61, 0x49, 0x00,
		0x69, 0x6f, 0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x00}

	v2 := T{I: 1, Io: 2, Iao: 0, R: 3}
	// {"I": 1, "Io": 2, "omitempty": 3}
	want2 := []byte{0xa3,
		0x61, 0x49, 0x01,
		0x62, 0x49, 0x6f, 0x02,
		0x69, 0x6f, 0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x03}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	tests := []roundTripTest{
		{"default values", v1, want1},
		{"non-default values", v2, want2}}
	testRoundTrip(t, tests, em, dm)
}

func TestOmitEmptyForBuiltinType(t *testing.T) {
	type T struct {
		B     bool           `cbor:"b"`
		Bo    bool           `cbor:"bo,omitempty"`
		UI    uint           `cbor:"ui"`
		UIo   uint           `cbor:"uio,omitempty"`
		I     int            `cbor:"i"`
		Io    int            `cbor:"io,omitempty"`
		F     float64        `cbor:"f"`
		Fo    float64        `cbor:"fo,omitempty"`
		S     string         `cbor:"s"`
		So    string         `cbor:"so,omitempty"`
		Slc   []string       `cbor:"slc"`
		Slco  []string       `cbor:"slco,omitempty"`
		M     map[int]string `cbor:"m"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		P     *int           `cbor:"p"`
		Po    *int           `cbor:"po,omitempty"`
		Intf  any            `cbor:"intf"`
		Intfo any            `cbor:"intfo,omitempty"`
	}

	v := T{}
	// {"b": false, "ui": 0, "i":0, "f": 0, "s": "", "slc": null, "m": {}, "p": nil, "intf": nil }
	want := []byte{0xa9,
		0x61, 0x62, 0xf4,
		0x62, 0x75, 0x69, 0x00,
		0x61, 0x69, 0x00,
		0x61, 0x66, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x73, 0x60,
		0x63, 0x73, 0x6c, 0x63, 0xf6,
		0x61, 0x6d, 0xf6,
		0x61, 0x70, 0xf6,
		0x64, 0x69, 0x6e, 0x74, 0x66, 0xf6,
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForAnonymousStruct(t *testing.T) {
	type T struct {
		Str  struct{} `cbor:"str"`
		Stro struct{} `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForStruct1(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForStruct2(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x65, 0x69, 0x6e, 0x74, 0x66, 0x6f, 0xf6} // {"stro": {intfo: nil}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"non-default values", v, want}}, em, dm)
}

func TestOmitEmptyForNestedStruct(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo,omitempty"`
	}
	type T2 struct {
		Stro T1 `cbor:"stro,omitempty"`
	}
	type T struct {
		Str  T2 `cbor:"str"`
		Stro T2 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForToArrayStruct1(t *testing.T) {
	type T1 struct {
		_    struct{} `cbor:",toarray"`
		b    bool
		ui   uint
		i    int
		f    float64
		s    string
		slc  []string
		m    map[int]string
		p    *int
		intf any
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{
		Str:  T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
		Stro: T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
	}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0x80} // {"str": []}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"no exportable fields", v, want}}, em, dm)
}

func TestOmitEmptyForToArrayStruct2(t *testing.T) {
	type T1 struct {
		_     struct{}       `cbor:",toarray"`
		Bo    bool           `cbor:"bo"`
		UIo   uint           `cbor:"uio"`
		Io    int            `cbor:"io"`
		Fo    float64        `cbor:"fo"`
		So    string         `cbor:"so"`
		Slco  []string       `cbor:"slco"`
		Mo    map[int]string `cbor:"mo"`
		Po    *int           `cbor:"po"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	// {"stro": [false, 0, 0, 0.0, "", [], {}, nil, nil]}
	want := []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0x89, 0xf4, 0x00, 0x00, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0xf6, 0xf6, 0xf6, 0xf6}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"has exportable fields", v, want}}, em, dm)
}

func TestOmitEmptyForStructWithPtrToAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitempty"`
			Y int `cbor:"y,omitempty"`
		}
		T2 struct {
			*T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitempty"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "null pointer to anonymous field",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not-null pointer to anonymous field",
			obj:          T{T2{&T1{}}},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{&T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{&T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{&T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitEmptyForStructWithAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitempty"`
			Y int `cbor:"y,omitempty"`
		}
		T2 struct {
			T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitempty"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "default values",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "default values",
			obj:          T{T2{T1{}}},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitEmptyForBinaryMarshaler1(t *testing.T) {
	type T1 struct {
		No number `cbor:"no,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	testCases := []roundTripTest{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestOmitEmptyForBinaryMarshaler2(t *testing.T) {
	type T1 struct {
		So stru `cbor:"so,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	testCases := []roundTripTest{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

// omitempty is a no-op for time.Time.
func TestOmitEmptyForTime(t *testing.T) {
	type T struct {
		Tm time.Time `cbor:"t,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x61, 0x74, 0xf6} // {"t": nil}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

// omitempty is a no-op for big.Int.
func TestOmitEmptyForBigInt(t *testing.T) {
	type T struct {
		I big.Int `cbor:"bi,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x62, 0x62, 0x69, 0xc2, 0x40} // {"bi": 2([])}

	em, _ := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTest{{"default values", v, want}}, em, dm)
}

func TestTaggedField(t *testing.T) {
	// A field (T2.X) with a tag dominates untagged field.
	type (
		T1 struct {
			S string
		}
		T2 struct {
			X string `cbor:"S"`
		}
		T struct {
			T1
			T2
		}
	)
	v := T{T1{"T1"}, T2{"T2"}}
	want := []byte{0xa1, 0x61, 0x53, 0x62, 0x54, 0x32} // {"S":"T2"}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{T2: T2{"T2"}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestDuplicatedFields(t *testing.T) {
	// Duplicate fields (T.T1.S, T.T2.S) are ignored.
	type (
		T1 struct {
			S string
		}
		T2 struct {
			S string
		}
		T3 struct {
			X string `cbor:"S"`
		}
		T4 struct {
			T1
			T3
		}
		T struct {
			T1
			T2
			T4 // Contains a tagged S field through T3; should not dominate.
		}
	)
	v := T{
		T1{"T1"},
		T2{"T2"},
		T4{
			T1{"nested T1"},
			T3{"nested T3"},
		},
	}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

type TReader struct {
	X int
}

func (s TReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func TestTaggedAnonymousField(t *testing.T) {
	// Anonymous field with a name given in its CBOR tag is treated as having that name, rather than being anonymous.
	type (
		T1 struct {
			X int
		}
		T struct {
			X  int
			T1 `cbor:"T1"`
		}
	)
	v := T{X: 1, T1: T1{X: 2}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x62, 0x54, 0x31, 0xa1, 0x61, 0x58, 0x02} // {X: 1, T1: {X:2}}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{X: 1, T1: T1{X: 2}}
	if err = Unmarshal(b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousInterfaceField(t *testing.T) {
	// Anonymous field of interface type is treated the same as having that type as its name, rather than being anonymous.
	type (
		T struct {
			X int
			io.Reader
		}
	)
	v := T{X: 1, Reader: TReader{X: 2}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x66, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0xa1, 0x61, 0x58, 0x02} // {X: 1, Reader: {X:2}}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	const wantErrorMsg = "cannot unmarshal map into Go struct field cbor.T.Reader of type io.Reader"
	if err = Unmarshal(b, &v2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", b)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", b, err)
		} else if !strings.Contains(typeError.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", b, err.Error(), wantErrorMsg)
		}
	}
}

func TestEncodeInterface(t *testing.T) {
	var r io.Reader = TReader{X: 2}
	want := []byte{0xa1, 0x61, 0x58, 0x02} // {X:2}
	b, err := Marshal(r)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", r, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", r, b, want)
	}

	var v io.Reader
	const wantErrorMsg = "cannot unmarshal map into Go value of type io.Reader"
	if err = Unmarshal(b, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", b)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", b, err)
		} else if !strings.Contains(typeError.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", b, err.Error(), wantErrorMsg)
		}
	}
}

func TestEncodeTime(t *testing.T) {
	timeUnixOpt := EncOptions{Time: TimeUnix}
	timeUnixMicroOpt := EncOptions{Time: TimeUnixMicro}
	timeUnixDynamicOpt := EncOptions{Time: TimeUnixDynamic}
	timeRFC3339Opt := EncOptions{Time: TimeRFC3339}
	timeRFC3339NanoOpt := EncOptions{Time: TimeRFC3339Nano}

	type timeConvert struct {
		opt          EncOptions
		wantCborData []byte
	}
	testCases := []struct {
		name    string
		tm      time.Time
		convert []timeConvert
	}{
		{
			name: "zero time",
			tm:   time.Time{},
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
			},
		},
		{
			name: "time without fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("1a514b67b0"), // 1363896240
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("fb41d452d9ec000000"), // 1363896240.0
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("1a514b67b0"), // 1363896240
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
			},
		},
		{
			name: "time with fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("1a514b67b0"), // 1363896240
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("fb41d452d9ec200000"), // 1363896240.5
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("fb41d452d9ec200000"), // 1363896240.5
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("76323031332d30332d32315432303a30343a30302e355a"), // "2013-03-21T20:04:00.5Z"
				},
			},
		},
		{
			name: "time before January 1, 1970 UTC without fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("3a0177f2cf"), // -24638160
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("fbc1777f2d00000000"), // -24638160.0
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("3a0177f2cf"), // -24638160
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("74313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("74313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
				},
			},
		},
	}
	for _, tc := range testCases {
		for _, convert := range tc.convert {
			var convertName string
			switch convert.opt.Time {
			case TimeUnix:
				convertName = "TimeUnix"
			case TimeUnixMicro:
				convertName = "TimeUnixMicro"
			case TimeUnixDynamic:
				convertName = "TimeUnixDynamic"
			case TimeRFC3339:
				convertName = "TimeRFC3339"
			case TimeRFC3339Nano:
				convertName = "TimeRFC3339Nano"
			}
			name := tc.name + " with " + convertName + " option"
			t.Run(name, func(t *testing.T) {
				em, err := convert.opt.EncMode()
				if err != nil {
					t.Errorf("EncMode() returned error %v", err)
				}
				b, err := em.Marshal(tc.tm)
				if err != nil {
					t.Errorf("Marshal(%+v) returned error %v", tc.tm, err)
				} else if !bytes.Equal(b, convert.wantCborData) {
					t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.tm, b, convert.wantCborData)
				}
			})
		}
	}
}

func TestEncodeTimeWithTag(t *testing.T) {
	timeUnixOpt := EncOptions{Time: TimeUnix, TimeTag: EncTagRequired}
	timeUnixMicroOpt := EncOptions{Time: TimeUnixMicro, TimeTag: EncTagRequired}
	timeUnixDynamicOpt := EncOptions{Time: TimeUnixDynamic, TimeTag: EncTagRequired}
	timeRFC3339Opt := EncOptions{Time: TimeRFC3339, TimeTag: EncTagRequired}
	timeRFC3339NanoOpt := EncOptions{Time: TimeRFC3339Nano, TimeTag: EncTagRequired}

	type timeConvert struct {
		opt          EncOptions
		wantCborData []byte
	}
	testCases := []struct {
		name    string
		tm      time.Time
		convert []timeConvert
	}{
		{
			name: "zero time",
			tm:   time.Time{},
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("f6"), // encode as CBOR null
				},
			},
		},
		{
			name: "time without fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("c11a514b67b0"), // 1363896240
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("c1fb41d452d9ec000000"), // 1363896240.0
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("c11a514b67b0"), // 1363896240
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("c074323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("c074323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
			},
		},
		{
			name: "time with fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("c11a514b67b0"), // 1363896240
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("c1fb41d452d9ec200000"), // 1363896240.5
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("c1fb41d452d9ec200000"), // 1363896240.5
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("c074323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("c076323031332d30332d32315432303a30343a30302e355a"), // "2013-03-21T20:04:00.5Z"
				},
			},
		},
		{
			name: "time before January 1, 1970 UTC without fractional seconds",
			tm:   parseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
			convert: []timeConvert{
				{
					opt:          timeUnixOpt,
					wantCborData: hexDecode("c13a0177f2cf"), // -24638160
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: hexDecode("c1fbc1777f2d00000000"), // -24638160.0
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: hexDecode("c13a0177f2cf"), // -24638160
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: hexDecode("c074313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: hexDecode("c074313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
				},
			},
		},
	}
	for _, tc := range testCases {
		for _, convert := range tc.convert {
			var convertName string
			switch convert.opt.Time {
			case TimeUnix:
				convertName = "TimeUnix"
			case TimeUnixMicro:
				convertName = "TimeUnixMicro"
			case TimeUnixDynamic:
				convertName = "TimeUnixDynamic"
			case TimeRFC3339:
				convertName = "TimeRFC3339"
			case TimeRFC3339Nano:
				convertName = "TimeRFC3339Nano"
			}
			name := tc.name + " with " + convertName + " option"
			t.Run(name, func(t *testing.T) {
				em, err := convert.opt.EncMode()
				if err != nil {
					t.Errorf("EncMode() returned error %v", err)
				}
				b, err := em.Marshal(tc.tm)
				if err != nil {
					t.Errorf("Marshal(%+v) returned error %v", tc.tm, err)
				} else if !bytes.Equal(b, convert.wantCborData) {
					t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.tm, b, convert.wantCborData)
				}
			})
		}
	}
}

func parseTime(layout string, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return tm
}

func TestInvalidTimeMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid TimeMode 100"
	_, err := EncOptions{Time: TimeMode(100)}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestMarshalStructTag1(t *testing.T) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := hexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag2(t *testing.T) {
	type strc struct {
		A string `json:"a"`
		B string `json:"b"`
		C string `json:"c"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := hexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag3(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"z"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := hexDecode("a36161614161626142617a6143") // {"a":"A", "b":"B", "z":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag4(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"-"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := hexDecode("a26161614161626142") // {"a":"A", "b":"B"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructLongFieldName(t *testing.T) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"abcdefghijklmnopqrstuvwxyz"`
		C string `cbor:"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := hexDecode("a361616141781a6162636465666768696a6b6c6d6e6f707172737475767778797a614278426162636465666768696a6b6c6d6e6f707172737475767778797a6162636465666768696a6b6c6d6e6f707172737475767778797a6162636465666768696a6b6c6d6e6143") // {"a":"A", "abcdefghijklmnopqrstuvwxyz":"B", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalRawMessageValue(t *testing.T) {
	type (
		T1 struct {
			M RawMessage `cbor:",omitempty"`
		}
		T2 struct {
			M *RawMessage `cbor:",omitempty"`
		}
	)

	var (
		rawNil   = RawMessage(nil)
		rawEmpty = RawMessage([]byte{})
		raw      = RawMessage([]byte{0x01})
	)

	tests := []struct {
		obj  any
		want []byte
	}{
		// Test with nil RawMessage.
		{rawNil, []byte{0xf6}},
		{&rawNil, []byte{0xf6}},
		{[]any{rawNil}, []byte{0x81, 0xf6}},
		{&[]any{rawNil}, []byte{0x81, 0xf6}},
		{[]any{&rawNil}, []byte{0x81, 0xf6}},
		{&[]any{&rawNil}, []byte{0x81, 0xf6}},
		{struct{ M RawMessage }{rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&struct{ M RawMessage }{rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{struct{ M *RawMessage }{&rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&struct{ M *RawMessage }{&rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{map[string]any{"M": rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&map[string]any{"M": rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{map[string]any{"M": &rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&map[string]any{"M": &rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{T1{rawNil}, []byte{0xa0}},
		{T2{&rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&T1{rawNil}, []byte{0xa0}},
		{&T2{&rawNil}, []byte{0xa1, 0x61, 0x4d, 0xf6}},

		// Test with empty, but non-nil, RawMessage.
		{rawEmpty, []byte{0xf6}},
		{&rawEmpty, []byte{0xf6}},
		{[]any{rawEmpty}, []byte{0x81, 0xf6}},
		{&[]any{rawEmpty}, []byte{0x81, 0xf6}},
		{[]any{&rawEmpty}, []byte{0x81, 0xf6}},
		{&[]any{&rawEmpty}, []byte{0x81, 0xf6}},
		{struct{ M RawMessage }{rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&struct{ M RawMessage }{rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{struct{ M *RawMessage }{&rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&struct{ M *RawMessage }{&rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{map[string]any{"M": rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&map[string]any{"M": rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{map[string]any{"M": &rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&map[string]any{"M": &rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{T1{rawEmpty}, []byte{0xa0}},
		{T2{&rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},
		{&T1{rawEmpty}, []byte{0xa0}},
		{&T2{&rawEmpty}, []byte{0xa1, 0x61, 0x4d, 0xf6}},

		// Test with RawMessage with some data.
		{raw, []byte{0x01}},
		{&raw, []byte{0x01}},
		{[]any{raw}, []byte{0x81, 0x01}},
		{&[]any{raw}, []byte{0x81, 0x01}},
		{[]any{&raw}, []byte{0x81, 0x01}},
		{&[]any{&raw}, []byte{0x81, 0x01}},
		{struct{ M RawMessage }{raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&struct{ M RawMessage }{raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{struct{ M *RawMessage }{&raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&struct{ M *RawMessage }{&raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{map[string]any{"M": raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&map[string]any{"M": raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{map[string]any{"M": &raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&map[string]any{"M": &raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{T1{raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{T2{&raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&T1{raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
		{&T2{&raw}, []byte{0xa1, 0x61, 0x4d, 0x01}},
	}

	for _, tc := range tests {
		b, err := Marshal(tc.obj)
		if err != nil {
			t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
		}
		if !bytes.Equal(b, tc.want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.want)
		}
	}
}

func TestCyclicDataStructure(t *testing.T) {
	type Node struct {
		V int   `cbor:"v"`
		N *Node `cbor:"n,omitempty"`
	}
	v := Node{1, &Node{2, &Node{3, nil}}}                                                                                  // linked list: 1, 2, 3
	wantCborData := []byte{0xa2, 0x61, 0x76, 0x01, 0x61, 0x6e, 0xa2, 0x61, 0x76, 0x02, 0x61, 0x6e, 0xa1, 0x61, 0x76, 0x03} // {v: 1, n: {v: 2, n: {v: 3}}}
	cborData, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, cborData) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, cborData, wantCborData)
	}
	var v1 Node
	if err = Unmarshal(cborData, &v1); err != nil {
		t.Fatalf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, v1) {
		t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", cborData, v1, v)
	}
}

func TestMarshalUnmarshalStructKeyAsInt(t *testing.T) {
	type T struct {
		F1 int `cbor:"1,omitempty,keyasint"`
		F2 int `cbor:"2,omitempty"`
		F3 int `cbor:"-3,omitempty,keyasint"`
	}
	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			"Zero value struct",
			T{},
			hexDecode("a0"), // {}
		},
		{
			"Initialized value struct",
			T{F1: 1, F2: 2, F3: 3},
			hexDecode("a301012203613202"), // {1: 1, -3: 3, "2": 2}
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			var v2 T
			if err := Unmarshal(b, &v2); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			}
			if !reflect.DeepEqual(tc.obj, v2) {
				t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
			}
		})
	}
}

func TestMarshalStructKeyAsIntNumError(t *testing.T) {
	type T1 struct {
		F1 int `cbor:"2.0,keyasint"`
	}
	type T2 struct {
		F1 int `cbor:"-18446744073709551616,keyasint"`
	}
	testCases := []struct {
		name         string
		obj          any
		wantErrorMsg string
	}{
		{
			name:         "float as key",
			obj:          T1{},
			wantErrorMsg: "cbor: failed to parse field name \"2.0\" to int",
		},
		{
			name:         "out of range int as key",
			obj:          T2{},
			wantErrorMsg: "cbor: failed to parse field name \"-18446744073709551616\" to int",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			switch {
			case err == nil:
				t.Errorf("Marshal(%+v) didn't return an error, want error %q", tc.obj, tc.wantErrorMsg)
			case !strings.Contains(err.Error(), tc.wantErrorMsg):
				t.Errorf("Marshal(%v) error %v, want %v", tc.obj, err.Error(), tc.wantErrorMsg)
			case b != nil:
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.obj, b)
			}
		})
	}
}

func TestMarshalUnmarshalStructToArray(t *testing.T) {
	type T1 struct {
		M int `cbor:",omitempty"`
	}
	type T2 struct {
		N int `cbor:",omitempty"`
		O int `cbor:",omitempty"`
	}
	type T struct {
		_   struct{} `cbor:",toarray"`
		A   int      `cbor:",omitempty"`
		B   T1       // nested struct
		T1           // embedded struct
		*T2          // embedded struct
	}
	testCases := []struct {
		name         string
		obj          T
		wantCborData []byte
	}{
		{
			"Zero value struct (test omitempty)",
			T{},
			hexDecode("8500a000f6f6"), // [0, {}, 0, nil, nil]
		},
		{
			"Initialized struct",
			T{A: 24, B: T1{M: 1}, T1: T1{M: 2}, T2: &T2{N: 3, O: 4}},
			hexDecode("851818a1614d01020304"), // [24, {M: 1}, 2, 3, 4]
		},
		{
			"Null pointer to embedded struct",
			T{A: 24, B: T1{M: 1}, T1: T1{M: 2}},
			hexDecode("851818a1614d0102f6f6"), // [24, {M: 1}, 2, nil, nil]
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			// SortMode should be ignored for struct to array encoding
			em, err := EncOptions{Sort: SortCanonical}.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err = em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			var v2 T
			if err := Unmarshal(b, &v2); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			}
			if tc.obj.T2 == nil {
				if !reflect.DeepEqual(*(v2.T2), T2{}) {
					t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
				}
				v2.T2 = nil
			}
			if !reflect.DeepEqual(tc.obj, v2) {
				t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
			}
		})
	}
}

func TestMapSort(t *testing.T) {
	m := make(map[any]bool)
	m[10] = true
	m[100] = true
	m[-1] = true
	m["z"] = true
	m["aa"] = true
	m[[1]int{100}] = true
	m[[1]int{-1}] = true
	m[false] = true

	lenFirstSortedCborData := hexDecode("a80af520f5f4f51864f5617af58120f5626161f5811864f5") // sorted keys: 10, -1, false, 100, "z", [-1], "aa", [100]
	bytewiseSortedCborData := hexDecode("a80af51864f520f5617af5626161f5811864f58120f5f4f5") // sorted keys: 10, 100, -1, "z", "aa", [100], [-1], false

	testCases := []struct {
		name         string
		opts         EncOptions
		wantCborData []byte
	}{
		{"Length first sort", EncOptions{Sort: SortLengthFirst}, lenFirstSortedCborData},
		{"Bytewise sort", EncOptions{Sort: SortBytewiseLexical}, bytewiseSortedCborData},
		{"CBOR canonical sort", EncOptions{Sort: SortCanonical}, lenFirstSortedCborData},
		{"CTAP2 canonical sort", EncOptions{Sort: SortCTAP2}, bytewiseSortedCborData},
		{"Core deterministic sort", EncOptions{Sort: SortCoreDeterministic}, bytewiseSortedCborData},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err := em.Marshal(m)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", m, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", m, b, tc.wantCborData)
			}
		})
	}
}

func TestStructSort(t *testing.T) {
	type T struct {
		A bool `cbor:"aa"`
		B bool `cbor:"z"`
		C bool `cbor:"-1,keyasint"`
		D bool `cbor:"100,keyasint"`
		E bool `cbor:"10,keyasint"`
	}
	var v T

	unsortedCborData := hexDecode("a5626161f4617af420f41864f40af4")       // unsorted fields: "aa", "z", -1, 100, 10
	lenFirstSortedCborData := hexDecode("a50af420f41864f4617af4626161f4") // sorted fields: 10, -1, 100, "z", "aa",
	bytewiseSortedCborData := hexDecode("a50af41864f420f4617af4626161f4") // sorted fields: 10, 100, -1, "z", "aa"

	testCases := []struct {
		name         string
		opts         EncOptions
		wantCborData []byte
	}{
		{"No sort", EncOptions{}, unsortedCborData},
		{"No sort", EncOptions{Sort: SortNone}, unsortedCborData},
		{"Length first sort", EncOptions{Sort: SortLengthFirst}, lenFirstSortedCborData},
		{"Bytewise sort", EncOptions{Sort: SortBytewiseLexical}, bytewiseSortedCborData},
		{"CBOR canonical sort", EncOptions{Sort: SortCanonical}, lenFirstSortedCborData},
		{"CTAP2 canonical sort", EncOptions{Sort: SortCTAP2}, bytewiseSortedCborData},
		{"Core deterministic sort", EncOptions{Sort: SortCoreDeterministic}, bytewiseSortedCborData},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err := em.Marshal(v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", v, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, tc.wantCborData)
			}
		})
	}
}

func TestInvalidSort(t *testing.T) {
	wantErrorMsg := "cbor: invalid SortMode 100"
	_, err := EncOptions{Sort: SortMode(100)}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestTypeAlias(t *testing.T) { //nolint:dupl,unconvert
	type myBool = bool
	type myUint = uint
	type myUint8 = uint8
	type myUint16 = uint16
	type myUint32 = uint32
	type myUint64 = uint64
	type myInt = int
	type myInt8 = int8
	type myInt16 = int16
	type myInt32 = int32
	type myInt64 = int64
	type myFloat32 = float32
	type myFloat64 = float64
	type myString = string
	type myByteSlice = []byte
	type myIntSlice = []int
	type myIntArray = [4]int
	type myMapIntInt = map[int]int

	testCases := []roundTripTest{
		{
			name:         "bool alias",
			obj:          myBool(true),
			wantCborData: hexDecode("f5"),
		},
		{
			name:         "uint alias",
			obj:          myUint(0),
			wantCborData: hexDecode("00"),
		},
		{
			name:         "uint8 alias",
			obj:          myUint8(0),
			wantCborData: hexDecode("00"),
		},
		{
			name:         "uint16 alias",
			obj:          myUint16(1000),
			wantCborData: hexDecode("1903e8"),
		},
		{
			name:         "uint32 alias",
			obj:          myUint32(1000000),
			wantCborData: hexDecode("1a000f4240"),
		},
		{
			name:         "uint64 alias",
			obj:          myUint64(1000000000000),
			wantCborData: hexDecode("1b000000e8d4a51000"),
		},
		{
			name:         "int alias",
			obj:          myInt(-1),
			wantCborData: hexDecode("20"),
		},
		{
			name:         "int8 alias",
			obj:          myInt8(-1),
			wantCborData: hexDecode("20"),
		},
		{
			name:         "int16 alias",
			obj:          myInt16(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "int32 alias",
			obj:          myInt32(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "int64 alias",
			obj:          myInt64(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "float32 alias",
			obj:          myFloat32(100000.0),
			wantCborData: hexDecode("fa47c35000"),
		},
		{
			name:         "float64 alias",
			obj:          myFloat64(1.1),
			wantCborData: hexDecode("fb3ff199999999999a"),
		},
		{
			name:         "string alias",
			obj:          myString("a"),
			wantCborData: hexDecode("6161"),
		},
		{
			name:         "[]byte alias",
			obj:          myByteSlice([]byte{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: hexDecode("4401020304"),
		},
		{
			name:         "[]int alias",
			obj:          myIntSlice([]int{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: hexDecode("8401020304"),
		},
		{
			name:         "[4]int alias",
			obj:          myIntArray([...]int{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: hexDecode("8401020304"),
		},
		{
			name:         "map[int]int alias",
			obj:          myMapIntInt(map[int]int{1: 2, 3: 4}), //nolint:unconvert
			wantCborData: hexDecode("a201020304"),
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	}
	testRoundTrip(t, testCases, em, dm)
}

func TestNewTypeWithBuiltinUnderlyingType(t *testing.T) { //nolint:dupl
	type myBool bool
	type myUint uint
	type myUint8 uint8
	type myUint16 uint16
	type myUint32 uint32
	type myUint64 uint64
	type myInt int
	type myInt8 int8
	type myInt16 int16
	type myInt32 int32
	type myInt64 int64
	type myFloat32 float32
	type myFloat64 float64
	type myString string
	type myByteSlice []byte
	type myIntSlice []int
	type myIntArray [4]int
	type myMapIntInt map[int]int

	testCases := []roundTripTest{
		{
			name:         "bool alias",
			obj:          myBool(true),
			wantCborData: hexDecode("f5"),
		},
		{
			name:         "uint alias",
			obj:          myUint(0),
			wantCborData: hexDecode("00"),
		},
		{
			name:         "uint8 alias",
			obj:          myUint8(0),
			wantCborData: hexDecode("00"),
		},
		{
			name:         "uint16 alias",
			obj:          myUint16(1000),
			wantCborData: hexDecode("1903e8"),
		},
		{
			name:         "uint32 alias",
			obj:          myUint32(1000000),
			wantCborData: hexDecode("1a000f4240"),
		},
		{
			name:         "uint64 alias",
			obj:          myUint64(1000000000000),
			wantCborData: hexDecode("1b000000e8d4a51000"),
		},
		{
			name:         "int alias",
			obj:          myInt(-1),
			wantCborData: hexDecode("20"),
		},
		{
			name:         "int8 alias",
			obj:          myInt8(-1),
			wantCborData: hexDecode("20"),
		},
		{
			name:         "int16 alias",
			obj:          myInt16(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "int32 alias",
			obj:          myInt32(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "int64 alias",
			obj:          myInt64(-1000),
			wantCborData: hexDecode("3903e7"),
		},
		{
			name:         "float32 alias",
			obj:          myFloat32(100000.0),
			wantCborData: hexDecode("fa47c35000"),
		},
		{
			name:         "float64 alias",
			obj:          myFloat64(1.1),
			wantCborData: hexDecode("fb3ff199999999999a"),
		},
		{
			name:         "string alias",
			obj:          myString("a"),
			wantCborData: hexDecode("6161"),
		},
		{
			name:         "[]byte alias",
			obj:          myByteSlice([]byte{1, 2, 3, 4}),
			wantCborData: hexDecode("4401020304"),
		},
		{
			name:         "[]int alias",
			obj:          myIntSlice([]int{1, 2, 3, 4}),
			wantCborData: hexDecode("8401020304"),
		},
		{
			name:         "[4]int alias",
			obj:          myIntArray([...]int{1, 2, 3, 4}),
			wantCborData: hexDecode("8401020304"),
		},
		{
			name:         "map[int]int alias",
			obj:          myMapIntInt(map[int]int{1: 2, 3: 4}),
			wantCborData: hexDecode("a201020304"),
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	}
	testRoundTrip(t, testCases, em, dm)
}

func TestShortestFloat16(t *testing.T) {
	testCases := []struct {
		name         string
		f64          float64
		wantCborData []byte
	}{
		// Data from RFC 7049 appendix A
		{"Shrink to float16", 0.0, hexDecode("f90000")},
		{"Shrink to float16", 1.0, hexDecode("f93c00")},
		{"Shrink to float16", 1.5, hexDecode("f93e00")},
		{"Shrink to float16", 65504.0, hexDecode("f97bff")},
		{"Shrink to float16", 5.960464477539063e-08, hexDecode("f90001")},
		{"Shrink to float16", 6.103515625e-05, hexDecode("f90400")},
		{"Shrink to float16", -4.0, hexDecode("f9c400")},
		// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{"Shrink to float16", 0.333251953125, hexDecode("f93555")},
		// Data from 7049bis 4.2.1 and 5.5
		{"Shrink to float16", 5.5, hexDecode("f94580")},
		// Data from RFC 7049 appendix A
		{"Shrink to float32", 100000.0, hexDecode("fa47c35000")},
		{"Shrink to float32", 3.4028234663852886e+38, hexDecode("fa7f7fffff")},
		// Data from 7049bis 4.2.1 and 5.5
		{"Shrink to float32", 5555.5, hexDecode("fa45ad9c00")},
		{"Shrink to float32", 1000000.5, hexDecode("fa49742408")},
		// Data from RFC 7049 appendix A
		{"Shrink to float64", 1.0e+300, hexDecode("fb7e37e43c8800759c")},
	}
	em, err := EncOptions{ShortestFloat: ShortestFloat16}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.f64)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.f64, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f64, b, tc.wantCborData)
			}
			var f64 float64
			if err = Unmarshal(b, &f64); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			} else if f64 != tc.f64 {
				t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f64)
			}
		})
	}
}

/*
	func TestShortestFloat32(t *testing.T) {
		testCases := []struct {
			name         string
			f64          float64
			wantCborData []byte
		}{
			// Data from RFC 7049 appendix A
			{"Shrink to float32", 0.0, hexDecode("fa00000000")},
			{"Shrink to float32", 1.0, hexDecode("fa3f800000")},
			{"Shrink to float32", 1.5, hexDecode("fa3fc00000")},
			{"Shrink to float32", 65504.0, hexDecode("fa477fe000")},
			{"Shrink to float32", 5.960464477539063e-08, hexDecode("fa33800000")},
			{"Shrink to float32", 6.103515625e-05, hexDecode("fa38800000")},
			{"Shrink to float32", -4.0, hexDecode("fac0800000")},
			// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
			{"Shrink to float32", 0.333251953125, hexDecode("fa3eaaa000")},
			// Data from 7049bis 4.2.1 and 5.5
			{"Shrink to float32", 5.5, hexDecode("fa40b00000")},
			// Data from RFC 7049 appendix A
			{"Shrink to float32", 100000.0, hexDecode("fa47c35000")},
			{"Shrink to float32", 3.4028234663852886e+38, hexDecode("fa7f7fffff")},
			// Data from 7049bis 4.2.1 and 5.5
			{"Shrink to float32", 5555.5, hexDecode("fa45ad9c00")},
			{"Shrink to float32", 1000000.5, hexDecode("fa49742408")},
			// Data from RFC 7049 appendix A
			{"Shrink to float64", 1.0e+300, hexDecode("fb7e37e43c8800759c")},
		}
		em, err := EncOptions{ShortestFloat: ShortestFloat32}.EncMode()
		if err != nil {
			t.Errorf("EncMode() returned an error %v", err)
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				b, err := em.Marshal(tc.f64)
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.f64, err)
				} else if !bytes.Equal(b, tc.wantCborData) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f64, b, tc.wantCborData)
				}
				var f64 float64
				if err = Unmarshal(b, &f64); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f64 != tc.f64 {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f64)
				}
			})
		}
	}

	func TestShortestFloat64(t *testing.T) {
		testCases := []struct {
			name         string
			f64          float64
			wantCborData []byte
		}{
			// Data from RFC 7049 appendix A
			{"Shrink to float64", 0.0, hexDecode("fb0000000000000000")},
			{"Shrink to float64", 1.0, hexDecode("fb3ff0000000000000")},
			{"Shrink to float64", 1.5, hexDecode("fb3ff8000000000000")},
			{"Shrink to float64", 65504.0, hexDecode("fb40effc0000000000")},
			{"Shrink to float64", 5.960464477539063e-08, hexDecode("fb3e70000000000000")},
			{"Shrink to float64", 6.103515625e-05, hexDecode("fb3f10000000000000")},
			{"Shrink to float64", -4.0, hexDecode("fbc010000000000000")},
			// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
			{"Shrink to float64", 0.333251953125, hexDecode("fb3fd5540000000000")},
			// Data from 7049bis 4.2.1 and 5.5
			{"Shrink to float64", 5.5, hexDecode("fb4016000000000000")},
			// Data from RFC 7049 appendix A
			{"Shrink to float64", 100000.0, hexDecode("fb40f86a0000000000")},
			{"Shrink to float64", 3.4028234663852886e+38, hexDecode("fb47efffffe0000000")},
			// Data from 7049bis 4.2.1 and 5.5
			{"Shrink to float64", 5555.5, hexDecode("fb40b5b38000000000")},
			{"Shrink to float64", 1000000.5, hexDecode("fb412e848100000000")},
			// Data from RFC 7049 appendix A
			{"Shrink to float64", 1.0e+300, hexDecode("fb7e37e43c8800759c")},
		}
		em, err := EncOptions{ShortestFloat: ShortestFloat64}.EncMode()
		if err != nil {
			t.Errorf("EncMode() returned an error %v", err)
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				b, err := em.Marshal(tc.f64)
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.f64, err)
				} else if !bytes.Equal(b, tc.wantCborData) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f64, b, tc.wantCborData)
				}
				var f64 float64
				if err = Unmarshal(b, &f64); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f64 != tc.f64 {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f64)
				}
			})
		}
	}
*/
func TestShortestFloatNone(t *testing.T) {
	testCases := []struct {
		name         string
		f            any
		wantCborData []byte
	}{
		// Data from RFC 7049 appendix A
		{"float32", float32(0.0), hexDecode("fa00000000")},
		{"float64", float64(0.0), hexDecode("fb0000000000000000")},
		{"float32", float32(1.0), hexDecode("fa3f800000")},
		{"float64", float64(1.0), hexDecode("fb3ff0000000000000")},
		{"float32", float32(1.5), hexDecode("fa3fc00000")},
		{"float64", float64(1.5), hexDecode("fb3ff8000000000000")},
		{"float32", float32(65504.0), hexDecode("fa477fe000")},
		{"float64", float64(65504.0), hexDecode("fb40effc0000000000")},
		{"float32", float32(5.960464477539063e-08), hexDecode("fa33800000")},
		{"float64", float64(5.960464477539063e-08), hexDecode("fb3e70000000000000")},
		{"float32", float32(6.103515625e-05), hexDecode("fa38800000")},
		{"float64", float64(6.103515625e-05), hexDecode("fb3f10000000000000")},
		{"float32", float32(-4.0), hexDecode("fac0800000")},
		{"float64", float64(-4.0), hexDecode("fbc010000000000000")},
		// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{"float32", float32(0.333251953125), hexDecode("fa3eaaa000")},
		{"float64", float64(0.333251953125), hexDecode("fb3fd5540000000000")},
		// Data from 7049bis 4.2.1 and 5.5
		{"float32", float32(5.5), hexDecode("fa40b00000")},
		{"float64", float64(5.5), hexDecode("fb4016000000000000")},
		// Data from RFC 7049 appendix A
		{"float32", float32(100000.0), hexDecode("fa47c35000")},
		{"float64", float64(100000.0), hexDecode("fb40f86a0000000000")},
		{"float32", float32(3.4028234663852886e+38), hexDecode("fa7f7fffff")},
		{"float64", float64(3.4028234663852886e+38), hexDecode("fb47efffffe0000000")},
		// Data from 7049bis 4.2.1 and 5.5
		{"float32", float32(5555.5), hexDecode("fa45ad9c00")},
		{"float64", float64(5555.5), hexDecode("fb40b5b38000000000")},
		{"float32", float32(1000000.5), hexDecode("fa49742408")},
		{"float64", float64(1000000.5), hexDecode("fb412e848100000000")},
		{"float64", float64(1.0e+300), hexDecode("fb7e37e43c8800759c")},
	}
	em, err := EncOptions{ShortestFloat: ShortestFloatNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.f)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.f, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f, b, tc.wantCborData)
			}
			if reflect.ValueOf(tc.f).Kind() == reflect.Float32 {
				var f32 float32
				if err = Unmarshal(b, &f32); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f32 != tc.f {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f32, tc.f)
				}
			} else {
				var f64 float64
				if err = Unmarshal(b, &f64); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f64 != tc.f {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f)
				}
			}
		})
	}
}

func TestInvalidShortestFloat(t *testing.T) {
	wantErrorMsg := "cbor: invalid ShortestFloatMode 100"
	_, err := EncOptions{ShortestFloat: ShortestFloatMode(100)}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestInfConvert(t *testing.T) {
	infConvertNoneOpt := EncOptions{InfConvert: InfConvertNone}
	infConvertFloat16Opt := EncOptions{InfConvert: InfConvertFloat16}
	testCases := []struct {
		name         string
		v            any
		opts         EncOptions
		wantCborData []byte
	}{
		{"float32 -inf no conversion", float32(math.Inf(-1)), infConvertNoneOpt, hexDecode("faff800000")},
		{"float32 +inf no conversion", float32(math.Inf(1)), infConvertNoneOpt, hexDecode("fa7f800000")},
		{"float64 -inf no conversion", math.Inf(-1), infConvertNoneOpt, hexDecode("fbfff0000000000000")},
		{"float64 +inf no conversion", math.Inf(1), infConvertNoneOpt, hexDecode("fb7ff0000000000000")},
		{"float32 -inf to float16", float32(math.Inf(-1)), infConvertFloat16Opt, hexDecode("f9fc00")},
		{"float32 +inf to float16", float32(math.Inf(1)), infConvertFloat16Opt, hexDecode("f97c00")},
		{"float64 -inf to float16", math.Inf(-1), infConvertFloat16Opt, hexDecode("f9fc00")},
		{"float64 +inf to float16", math.Inf(1), infConvertFloat16Opt, hexDecode("f97c00")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned an error %v", err)
			}
			b, err := em.Marshal(tc.v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.v, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.v, b, tc.wantCborData)
			}
		})
	}
}

func TestInvalidInfConvert(t *testing.T) {
	wantErrorMsg := "cbor: invalid InfConvertMode 100"
	_, err := EncOptions{InfConvert: InfConvertMode(100)}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

// Keith Randall's workaround for constant propagation issue https://github.com/golang/go/issues/36400
const (
	// qnan 32 bits constants
	qnanConst0xffc00001 uint32 = 0xffc00001
	qnanConst0x7fc00001 uint32 = 0x7fc00001
	qnanConst0xffc02000 uint32 = 0xffc02000
	qnanConst0x7fc02000 uint32 = 0x7fc02000
	// snan 32 bits constants
	snanConst0xff800001 uint32 = 0xff800001
	snanConst0x7f800001 uint32 = 0x7f800001
	snanConst0xff802000 uint32 = 0xff802000
	snanConst0x7f802000 uint32 = 0x7f802000
	// qnan 64 bits constants
	qnanConst0xfff8000000000001 uint64 = 0xfff8000000000001
	qnanConst0x7ff8000000000001 uint64 = 0x7ff8000000000001
	qnanConst0xfff8000020000000 uint64 = 0xfff8000020000000
	qnanConst0x7ff8000020000000 uint64 = 0x7ff8000020000000
	qnanConst0xfffc000000000000 uint64 = 0xfffc000000000000
	qnanConst0x7ffc000000000000 uint64 = 0x7ffc000000000000
	// snan 64 bits constants
	snanConst0xfff0000000000001 uint64 = 0xfff0000000000001
	snanConst0x7ff0000000000001 uint64 = 0x7ff0000000000001
	snanConst0xfff0000020000000 uint64 = 0xfff0000020000000
	snanConst0x7ff0000020000000 uint64 = 0x7ff0000020000000
	snanConst0xfff4000000000000 uint64 = 0xfff4000000000000
	snanConst0x7ff4000000000000 uint64 = 0x7ff4000000000000
)

var (
	// qnan 32 bits variables
	qnanVar0xffc00001 = qnanConst0xffc00001
	qnanVar0x7fc00001 = qnanConst0x7fc00001
	qnanVar0xffc02000 = qnanConst0xffc02000
	qnanVar0x7fc02000 = qnanConst0x7fc02000
	// snan 32 bits variables
	snanVar0xff800001 = snanConst0xff800001
	snanVar0x7f800001 = snanConst0x7f800001
	snanVar0xff802000 = snanConst0xff802000
	snanVar0x7f802000 = snanConst0x7f802000
	// qnan 64 bits variables
	qnanVar0xfff8000000000001 = qnanConst0xfff8000000000001
	qnanVar0x7ff8000000000001 = qnanConst0x7ff8000000000001
	qnanVar0xfff8000020000000 = qnanConst0xfff8000020000000
	qnanVar0x7ff8000020000000 = qnanConst0x7ff8000020000000
	qnanVar0xfffc000000000000 = qnanConst0xfffc000000000000
	qnanVar0x7ffc000000000000 = qnanConst0x7ffc000000000000
	// snan 64 bits variables
	snanVar0xfff0000000000001 = snanConst0xfff0000000000001
	snanVar0x7ff0000000000001 = snanConst0x7ff0000000000001
	snanVar0xfff0000020000000 = snanConst0xfff0000020000000
	snanVar0x7ff0000020000000 = snanConst0x7ff0000020000000
	snanVar0xfff4000000000000 = snanConst0xfff4000000000000
	snanVar0x7ff4000000000000 = snanConst0x7ff4000000000000
)

func TestNaNConvert(t *testing.T) {
	nanConvert7e00Opt := EncOptions{NaNConvert: NaNConvert7e00}
	nanConvertNoneOpt := EncOptions{NaNConvert: NaNConvertNone}
	nanConvertPreserveSignalOpt := EncOptions{NaNConvert: NaNConvertPreserveSignal}
	nanConvertQuietOpt := EncOptions{NaNConvert: NaNConvertQuiet}

	type nanConvert struct {
		opt          EncOptions
		wantCborData []byte
	}
	testCases := []struct {
		v       any
		convert []nanConvert
	}{
		// float32 qNaN dropped payload not zero
		{math.Float32frombits(qnanVar0xffc00001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("faffc00001")},
			{nanConvertPreserveSignalOpt, hexDecode("faffc00001")},
			{nanConvertQuietOpt, hexDecode("faffc00001")},
		}},
		// float32 qNaN dropped payload not zero
		{math.Float32frombits(qnanVar0x7fc00001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fa7fc00001")},
			{nanConvertPreserveSignalOpt, hexDecode("fa7fc00001")},
			{nanConvertQuietOpt, hexDecode("fa7fc00001")},
		}},
		// float32 -qNaN dropped payload zero
		{math.Float32frombits(qnanVar0xffc02000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("faffc02000")},
			{nanConvertPreserveSignalOpt, hexDecode("f9fe01")},
			{nanConvertQuietOpt, hexDecode("f9fe01")},
		}},
		// float32 qNaN dropped payload zero
		{math.Float32frombits(qnanVar0x7fc02000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fa7fc02000")},
			{nanConvertPreserveSignalOpt, hexDecode("f97e01")},
			{nanConvertQuietOpt, hexDecode("f97e01")},
		}},
		// float32 -sNaN dropped payload not zero
		{math.Float32frombits(snanVar0xff800001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("faff800001")},
			{nanConvertPreserveSignalOpt, hexDecode("faff800001")},
			{nanConvertQuietOpt, hexDecode("faffc00001")},
		}},
		// float32 sNaN dropped payload not zero
		{math.Float32frombits(snanVar0x7f800001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fa7f800001")},
			{nanConvertPreserveSignalOpt, hexDecode("fa7f800001")},
			{nanConvertQuietOpt, hexDecode("fa7fc00001")},
		}},
		// float32 -sNaN dropped payload zero
		{math.Float32frombits(snanVar0xff802000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("faff802000")},
			{nanConvertPreserveSignalOpt, hexDecode("f9fc01")},
			{nanConvertQuietOpt, hexDecode("f9fe01")},
		}},
		// float32 sNaN dropped payload zero
		{math.Float32frombits(snanVar0x7f802000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fa7f802000")},
			{nanConvertPreserveSignalOpt, hexDecode("f97c01")},
			{nanConvertQuietOpt, hexDecode("f97e01")},
		}},
		// float64 -qNaN dropped payload not zero
		{math.Float64frombits(qnanVar0xfff8000000000001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfff8000000000001")},
			{nanConvertPreserveSignalOpt, hexDecode("fbfff8000000000001")},
			{nanConvertQuietOpt, hexDecode("fbfff8000000000001")},
		}},
		// float64 qNaN dropped payload not zero
		{math.Float64frombits(qnanVar0x7ff8000000000001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ff8000000000001")},
			{nanConvertPreserveSignalOpt, hexDecode("fb7ff8000000000001")},
			{nanConvertQuietOpt, hexDecode("fb7ff8000000000001")},
		}},
		// float64 -qNaN dropped payload zero
		{math.Float64frombits(qnanVar0xfff8000020000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfff8000020000000")},
			{nanConvertPreserveSignalOpt, hexDecode("faffc00001")},
			{nanConvertQuietOpt, hexDecode("faffc00001")},
		}},
		// float64 qNaN dropped payload zero
		{math.Float64frombits(qnanVar0x7ff8000020000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ff8000020000000")},
			{nanConvertPreserveSignalOpt, hexDecode("fa7fc00001")},
			{nanConvertQuietOpt, hexDecode("fa7fc00001")},
		}},
		// float64 -qNaN dropped payload zero
		{math.Float64frombits(qnanVar0xfffc000000000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfffc000000000000")},
			{nanConvertPreserveSignalOpt, hexDecode("f9ff00")},
			{nanConvertQuietOpt, hexDecode("f9ff00")},
		}},
		// float64 qNaN dropped payload zero
		{math.Float64frombits(qnanVar0x7ffc000000000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ffc000000000000")},
			{nanConvertPreserveSignalOpt, hexDecode("f97f00")},
			{nanConvertQuietOpt, hexDecode("f97f00")},
		}},
		// float64 -sNaN dropped payload not zero
		{math.Float64frombits(snanVar0xfff0000000000001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfff0000000000001")},
			{nanConvertPreserveSignalOpt, hexDecode("fbfff0000000000001")},
			{nanConvertQuietOpt, hexDecode("fbfff8000000000001")},
		}},
		// float64 sNaN dropped payload not zero
		{math.Float64frombits(snanVar0x7ff0000000000001), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ff0000000000001")},
			{nanConvertPreserveSignalOpt, hexDecode("fb7ff0000000000001")},
			{nanConvertQuietOpt, hexDecode("fb7ff8000000000001")},
		}},
		// float64 -sNaN dropped payload zero
		{math.Float64frombits(snanVar0xfff0000020000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfff0000020000000")},
			{nanConvertPreserveSignalOpt, hexDecode("faff800001")},
			{nanConvertQuietOpt, hexDecode("faffc00001")},
		}},
		// float64 sNaN dropped payload zero
		{math.Float64frombits(snanVar0x7ff0000020000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ff0000020000000")},
			{nanConvertPreserveSignalOpt, hexDecode("fa7f800001")},
			{nanConvertQuietOpt, hexDecode("fa7fc00001")},
		}},
		// float64 -sNaN dropped payload zero
		{math.Float64frombits(snanVar0xfff4000000000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fbfff4000000000000")},
			{nanConvertPreserveSignalOpt, hexDecode("f9fd00")},
			{nanConvertQuietOpt, hexDecode("f9ff00")},
		}},
		// float64 sNaN dropped payload zero
		{math.Float64frombits(snanVar0x7ff4000000000000), []nanConvert{
			{nanConvert7e00Opt, hexDecode("f97e00")},
			{nanConvertNoneOpt, hexDecode("fb7ff4000000000000")},
			{nanConvertPreserveSignalOpt, hexDecode("f97d00")},
			{nanConvertQuietOpt, hexDecode("f97f00")},
		}},
	}
	for _, tc := range testCases {
		for _, convert := range tc.convert {
			var convertName string
			switch convert.opt.NaNConvert {
			case NaNConvert7e00:
				convertName = "Convert7e00"
			case NaNConvertNone:
				convertName = "ConvertNone"
			case NaNConvertPreserveSignal:
				convertName = "ConvertPreserveSignal"
			case NaNConvertQuiet:
				convertName = "ConvertQuiet"
			}
			var vName string
			switch v := tc.v.(type) {
			case float32:
				vName = fmt.Sprintf("0x%x", math.Float32bits(v))
			case float64:
				vName = fmt.Sprintf("0x%x", math.Float64bits(v))
			}
			name := convertName + "_" + vName
			t.Run(name, func(t *testing.T) {
				em, err := convert.opt.EncMode()
				if err != nil {
					t.Errorf("EncMode() returned an error %v", err)
				}
				b, err := em.Marshal(tc.v)
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.v, err)
				} else if !bytes.Equal(b, convert.wantCborData) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.v, b, convert.wantCborData)
				}
			})
		}
	}
}

func TestInvalidNaNConvert(t *testing.T) {
	wantErrorMsg := "cbor: invalid NaNConvertMode 100"
	_, err := EncOptions{NaNConvert: NaNConvertMode(100)}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestMarshalSenML(t *testing.T) {
	// Data from https://tools.ietf.org/html/rfc8428#section-6
	// Data contains 13 floating-point numbers.
	cborData := hexDecode("87a721781b75726e3a6465763a6f773a3130653230373361303130383030363a22fb41d303a15b00106223614120050067766f6c7461676501615602fb405e066666666666a3006763757272656e74062402fb3ff3333333333333a3006763757272656e74062302fb3ff4cccccccccccda3006763757272656e74062202fb3ff6666666666666a3006763757272656e74062102f93e00a3006763757272656e74062002fb3ff999999999999aa3006763757272656e74060002fb3ffb333333333333")
	testCases := []struct {
		name string
		opts EncOptions
	}{
		{"EncOptions ShortestFloatNone", EncOptions{}},
		{"EncOptions ShortestFloat16", EncOptions{ShortestFloat: ShortestFloat16}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v []SenMLRecord
			if err := Unmarshal(cborData, &v); err != nil {
				t.Errorf("Marshal() returned error %v", err)
			}
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned an error %v", err)
			}
			b, err := em.Marshal(v)
			if err != nil {
				t.Errorf("Unmarshal() returned error %v ", err)
			}
			var v2 []SenMLRecord
			if err := Unmarshal(b, &v2); err != nil {
				t.Errorf("Marshal() returned error %v", err)
			}
			if !reflect.DeepEqual(v, v2) {
				t.Errorf("SenML round-trip failed: v1 %+v, v2 %+v", v, v2)
			}
		})
	}
}

func TestCanonicalEncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCanonical
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em, err := CanonicalEncOptions().EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CanonicalEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CanonicalEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CanonicalEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CanonicalEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestCTAP2EncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCTAP2
	wantShortestFloat := ShortestFloatNone
	wantNaNConvert := NaNConvertNone
	wantInfConvert := InfConvertNone
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em, err := CTAP2EncOptions().EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CTAP2EncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CTAP2EncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CTAP2EncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CTAP2EncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestCoreDetEncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCoreDeterministic
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em, err := CoreDetEncOptions().EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CoreDetEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CoreDetEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CoreDetEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CoreDetEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestPreferredUnsortedEncOptions(t *testing.T) {
	wantSortMode := SortNone
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	em, err := PreferredUnsortedEncOptions().EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err != nil {
		t.Errorf("StartIndefiniteArray() returned error %v", err)
	}
}

func TestEncModeInvalidIndefiniteLengthMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid IndefLength 101"
	_, err := EncOptions{IndefLength: 101}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncModeInvalidTagsMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid TagsMd 101"
	_, err := EncOptions{TagsMd: 101}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncModeInvalidBigIntConvertMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid BigIntConvertMode 101"
	_, err := EncOptions{BigIntConvert: 101}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncOptions(t *testing.T) {
	opts1 := EncOptions{
		Sort:          SortBytewiseLexical,
		ShortestFloat: ShortestFloat16,
		NaNConvert:    NaNConvertPreserveSignal,
		InfConvert:    InfConvertNone,
		BigIntConvert: BigIntConvertNone,
		Time:          TimeRFC3339Nano,
		TimeTag:       EncTagRequired,
		IndefLength:   IndefLengthForbidden,
		TagsMd:        TagsAllowed,
	}
	em, err := opts1.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	} else {
		opts2 := em.EncOptions()
		if !reflect.DeepEqual(opts1, opts2) {
			t.Errorf("EncOptions->EncMode->EncOptions returned different values: %v, %v", opts1, opts2)
		}
	}
}

func TestEncModeInvalidTimeTag(t *testing.T) {
	wantErrorMsg := "cbor: invalid TimeTag 100"
	_, err := EncOptions{TimeTag: 100}.EncMode()
	if err == nil {
		t.Errorf("EncMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncIndefiniteLengthOption(t *testing.T) {
	// Default option allows indefinite length items
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.StartIndefiniteByteString(); err != nil {
		t.Errorf("StartIndefiniteByteString() returned an error %v", err)
	}
	if err := enc.StartIndefiniteTextString(); err != nil {
		t.Errorf("StartIndefiniteTextString() returned an error %v", err)
	}
	if err := enc.StartIndefiniteArray(); err != nil {
		t.Errorf("StartIndefiniteArray() returned an error %v", err)
	}
	if err := enc.StartIndefiniteMap(); err != nil {
		t.Errorf("StartIndefiniteMap() returned an error %v", err)
	}

	// StartIndefiniteXXX returns error when IndefLength = IndefLengthForbidden
	em, _ := EncOptions{IndefLength: IndefLengthForbidden}.EncMode()
	enc = em.NewEncoder(&buf)
	wantErrorMsg := "cbor: indefinite-length byte string isn't allowed"
	if err := enc.StartIndefiniteByteString(); err == nil {
		t.Errorf("StartIndefiniteByteString() didn't return an error")
	} else if _, ok := err.(*IndefiniteLengthError); !ok {
		t.Errorf("StartIndefiniteByteString() error type %T, want *IndefiniteLengthError", err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteByteString() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	wantErrorMsg = "cbor: indefinite-length UTF-8 text string isn't allowed"
	if err := enc.StartIndefiniteTextString(); err == nil {
		t.Errorf("StartIndefiniteTextString() didn't return an error")
	} else if _, ok := err.(*IndefiniteLengthError); !ok {
		t.Errorf("StartIndefiniteTextString() error type %T, want *IndefiniteLengthError", err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteTextString() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	wantErrorMsg = "cbor: indefinite-length array isn't allowed"
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if _, ok := err.(*IndefiniteLengthError); !ok {
		t.Errorf("StartIndefiniteArray() error type %T, want *IndefiniteLengthError", err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	wantErrorMsg = "cbor: indefinite-length map isn't allowed"
	if err := enc.StartIndefiniteMap(); err == nil {
		t.Errorf("StartIndefiniteMap() didn't return an error")
	} else if _, ok := err.(*IndefiniteLengthError); !ok {
		t.Errorf("StartIndefiniteMap() error type %T, want *IndefiniteLengthError", err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteMap() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncTagsMdOption(t *testing.T) {
	// Default option allows encoding CBOR tags
	tag := Tag{123, "hello"}
	if _, err := Marshal(tag); err != nil {
		t.Errorf("Marshal() returned an error %v", err)
	}

	// Create EncMode with TimeTag = EncTagRequired and TagsForbidden option returns error
	wantErrorMsg := "cbor: cannot set TagsMd to TagsForbidden when TimeTag is EncTagRequired"
	_, err := EncOptions{TimeTag: EncTagRequired, TagsMd: TagsForbidden}.EncMode()
	if err == nil {
		t.Errorf("EncModeWithTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Create EncMode with TagSet and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot create EncMode with TagSet when TagsMd is TagsForbidden"
	tags := NewTagSet()
	_, err = EncOptions{TagsMd: TagsForbidden}.EncModeWithTags(tags)
	if err == nil {
		t.Errorf("EncModeWithTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	_, err = EncOptions{TagsMd: TagsForbidden}.EncModeWithSharedTags(tags)
	if err == nil {
		t.Errorf("EncModeWithSharedTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithSharedTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Encoding Tag and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot encode cbor.Tag when TagsMd is TagsForbidden"
	em, _ := EncOptions{TagsMd: TagsForbidden}.EncMode()
	if _, err := em.Marshal(&tag); err == nil {
		t.Errorf("Marshal() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Marshal() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Encoding RawTag and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot encode cbor.RawTag when TagsMd is TagsForbidden"
	rawTag := RawTag{123, []byte{01}}
	if _, err := em.Marshal(&rawTag); err == nil {
		t.Errorf("Marshal() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Marshal() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestMarshalPosBigInt(t *testing.T) {
	testCases := []struct {
		name             string
		cborDataShortest []byte
		cborDataBigInt   []byte
		value            big.Int
	}{
		{
			name:             "fit uint8",
			cborDataShortest: hexDecode("00"),
			cborDataBigInt:   hexDecode("c240"),
			value:            bigIntOrPanic("0"),
		},
		{
			name:             "fit uint16",
			cborDataShortest: hexDecode("1903e8"),
			cborDataBigInt:   hexDecode("c24203e8"),
			value:            bigIntOrPanic("1000"),
		},
		{
			name:             "fit uint32",
			cborDataShortest: hexDecode("1a000f4240"),
			cborDataBigInt:   hexDecode("c2430f4240"),
			value:            bigIntOrPanic("1000000"),
		},
		{
			name:             "fit uint64",
			cborDataShortest: hexDecode("1b000000e8d4a51000"),
			cborDataBigInt:   hexDecode("c245e8d4a51000"),
			value:            bigIntOrPanic("1000000000000"),
		},
		{
			name:             "max uint64",
			cborDataShortest: hexDecode("1bffffffffffffffff"),
			cborDataBigInt:   hexDecode("c248ffffffffffffffff"),
			value:            bigIntOrPanic("18446744073709551615"),
		},
		{
			name:             "overflow uint64",
			cborDataShortest: hexDecode("c249010000000000000000"),
			cborDataBigInt:   hexDecode("c249010000000000000000"),
			value:            bigIntOrPanic("18446744073709551616"),
		},
	}

	dmShortest, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmBigInt, err := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if b, err := dmShortest.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataShortest) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
			}

			if b, err := dmBigInt.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataBigInt) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataBigInt)
			}
		})
	}
}

func TestMarshalNegBigInt(t *testing.T) {
	testCases := []struct {
		name             string
		cborDataShortest []byte
		cborDataBigInt   []byte
		value            big.Int
	}{
		{
			name:             "fit int8",
			cborDataShortest: hexDecode("20"),
			cborDataBigInt:   hexDecode("c340"),
			value:            bigIntOrPanic("-1"),
		},
		{
			name:             "fit int16",
			cborDataShortest: hexDecode("3903e7"),
			cborDataBigInt:   hexDecode("c34203e7"),
			value:            bigIntOrPanic("-1000"),
		},
		{
			name:             "fit int32",
			cborDataShortest: hexDecode("3a000f423f"),
			cborDataBigInt:   hexDecode("c3430f423f"),
			value:            bigIntOrPanic("-1000000"),
		},
		{
			name:             "fit int64",
			cborDataShortest: hexDecode("3b000000e8d4a50fff"),
			cborDataBigInt:   hexDecode("c345e8d4a50fff"),
			value:            bigIntOrPanic("-1000000000000"),
		},
		{
			name:             "min int64",
			cborDataShortest: hexDecode("3b7fffffffffffffff"),
			cborDataBigInt:   hexDecode("c3487fffffffffffffff"),
			value:            bigIntOrPanic("-9223372036854775808"),
		},
		{
			name:             "overflow Go int64 fit CBOR neg int",
			cborDataShortest: hexDecode("3b8000000000000000"),
			cborDataBigInt:   hexDecode("c3488000000000000000"),
			value:            bigIntOrPanic("-9223372036854775809"),
		},
		{
			name:             "min CBOR neg int",
			cborDataShortest: hexDecode("3bffffffffffffffff"),
			cborDataBigInt:   hexDecode("c348ffffffffffffffff"),
			value:            bigIntOrPanic("-18446744073709551616"),
		},
		{
			name:             "overflow CBOR neg int",
			cborDataShortest: hexDecode("c349010000000000000000"),
			cborDataBigInt:   hexDecode("c349010000000000000000"),
			value:            bigIntOrPanic("-18446744073709551617"),
		},
	}

	dmShortest, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmBigInt, err := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if b, err := dmShortest.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataShortest) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
			}

			if b, err := dmBigInt.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataBigInt) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataBigInt)
			}
		})
	}
}

func TestMarshalByteStringUnwrap(t *testing.T) {
	testCases := []struct {
		name         string
		value        any
		wantCborData []byte
	}{
		{
			name: "map with ByteString keys",
			value: map[any]any{
				NewByteString(hexDecode("abcdef")): uint64(123),
			},
			wantCborData: hexDecode("A143ABCDEF187B"),
		},
	}
	dm, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("tc.value = %#v\n", tc.value)
			if b, err := dm.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.wantCborData)
			}
		})
	}
}
