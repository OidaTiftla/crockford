// Package crockford implements the Crockford base 32 encoding
//
// See https://www.crockford.com/base32.html
package crockford

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base32"
	"time"
)

// Base32 alphabets
const (
	LowercaseAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"
	UppercaseAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	UppercaseChecksum = UppercaseAlphabet + "*~$=U"
	LowercaseChecksum = LowercaseAlphabet + "*~$=u"
)

// Base32 encodings
var (
	Lower = base32.NewEncoding(LowercaseAlphabet).WithPadding(base32.NoPadding)
	Upper = base32.NewEncoding(UppercaseAlphabet).WithPadding(base32.NoPadding)
)

// Buffer lengths
const (
	LenTime   = 8  // length returned by AppendTime
	LenRandom = 8  // length returned by AppendRandom
	LenMD5    = 26 // length returned by AppendMD5
)

// Time encodes the Unix time as a 40-bit number. The resulting string is big endian
// and suitable for lexicographic sorting.
func Time(e *base32.Encoding, t time.Time) string {
	return string(AppendTime(e, t, nil))
}

// AppendTime appends onto dst LenTime bytes with the Unix time encoded as a 40-bit number.
// The resulting slice is big endian and suitable for lexicographic sorting.
func AppendTime(e *base32.Encoding, t time.Time, dst []byte) []byte {
	ut := t.Unix()
	var src [5]byte
	src[0] = byte(ut >> 32)
	src[1] = byte(ut >> 24)
	src[2] = byte(ut >> 16)
	src[3] = byte(ut >> 8)
	src[4] = byte(ut)
	ret, tar := ensure(LenTime, dst)
	e.Encode(tar, src[:])
	return ret
}

// mod calculates the big endian modulus of the byte string
func mod(b []byte, m int) (rem int) {
	for _, c := range b {
		rem = (rem*1<<8 + int(c)) % m
	}
	return
}

// Checksum returns the checksum byte for an unencoded body.
func Checksum(body []byte, uppercase bool) byte {
	alphabet := LowercaseChecksum
	if uppercase {
		alphabet = UppercaseChecksum
	}
	return alphabet[mod(body, 37)]
}

func normUpper(c byte) byte {
	switch c {
	case '0', 'O', 'o':
		return '0'
	case '1', 'I', 'i':
		return '1'
	case '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z', '*', '~', '$', '=', 'U':
		return c
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z', 'u':
		return c + 'A' - 'a'
	}
	return 0
}

// Normalized returns a normalized version of Crockford encoded bytes of src
// onto dst and returns the resulting slice. It replaces I with 1, o with 0,
// and removes invalid characters such as hyphens. The resulting slice is uppercase.
func Normalized(s string) string {
	return string(AppendNormalized(nil, []byte(s)))
}

// AppendNormalized appends a normalized version of Crockford encoded bytes of src
// onto dst and returns the resulting slice. It replaces I with 1, o with 0,
// and removes invalid characters such as hyphens. The resulting slice is uppercase.
func AppendNormalized(dst, src []byte) []byte {
	if cap(dst) == 0 {
		dst = make([]byte, 0, len(src))
	}
	for _, c := range src {
		if r := normUpper(c); r != 0 {
			dst = append(dst, r)
		}
	}
	return dst
}

// Random returns LenRandom (8) encoded bytes generated by crypto/rand.
func Random(e *base32.Encoding) string {
	return string(AppendRandom(e, nil))
}

// AppendRandom appends LenRandom (8) encoded bytes generated by crypto/rand onto dst.
func AppendRandom(e *base32.Encoding, dst []byte) []byte {
	// 5 bytes -> 8 base32 characters
	dst = grow(dst, LenRandom)
	src := dst[len(dst) : len(dst)+5]
	if _, err := rand.Read(src); err != nil {
		panic(err)
	}
	tar := dst[len(dst) : len(dst)+LenRandom]
	e.Encode(tar, src)
	return dst[:len(dst)+LenRandom]
}

// MD5 returns encoded bytes generated by MD5 hashing src.
func MD5(e *base32.Encoding, src []byte) string {
	return string(AppendMD5(e, nil, src))
}

// AppendMD5 appends LenMD (26) encoded bytes generated by MD5 hashing src onto dst.
func AppendMD5(e *base32.Encoding, dst, src []byte) []byte {
	//16 bytes -> 26 base32 characters
	var buf [md5.Size]byte

	h := md5.New()
	h.Write(src)
	h.Sum(buf[:0])

	// Ensure dst has 26 bytes capacity
	ret, tar := ensure(LenMD5, dst)
	e.Encode(tar, buf[:])
	return ret
}

func ensure(size int, b []byte) (ret, tar []byte) {
	ret = b
	newLen := len(b) + size
	if cap(b) >= newLen {
		ret = b[:newLen]
	} else {
		ret = append(b, make([]byte, size)...)
	}
	tar = ret[len(b):newLen]
	return
}

func grow(b []byte, n int) []byte {
	if cap(b)-len(b) > n {
		return b
	}
	return append(b, make([]byte, n)...)[:len(b)]
}
