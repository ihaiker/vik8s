package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	// Prefix is the identifier for the Apache-specific MD5 algorithm.
	Prefix = "$apr1$"

	// Size is the size of an MD5 checksum in bytes.
	Size = 16

	// Blocksize is the blocksize of APR1 in bytes.
	Blocksize = 64

	// Rounds is the number of rounds in the big loop.
	Rounds = 1000

	// validChars is used to create a base64-like string.
	validChars = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	numValidChars = big.NewInt(int64(len(validChars)))
)

// generateSalt generates a "base64 encoded" salt for apr1-md5 hash.
// This returns (nil, error) if no suitable source of randomness is found.
func generateSalt() []byte {
	salt := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res, err := rand.Int(rand.Reader, numValidChars)
		if err != nil {
			return nil
		}
		salt[i] = validChars[res.Uint64()]
	}
	return salt
}

func HashApr1(password string) (string, error) {
	return HashApr1WithSalt(password, string(generateSalt()))
}

// Hash hashes the given password, along with the salt.
// I did not design this algorithm, only re-implement and optimize slightly.
// This assumes strings are valid UTF-8 encoded strings.
// The salt must be 8 bytes long.
// This algorithm is adopted from the Java implementation found here:
// http://commons.apache.org/proper/commons-codec/apidocs/src-html/org/apache/commons/codec/digest/Md5Crypt.html
func HashApr1WithSalt(password, salt string) (string, error) {
	pwBytes := []byte(password)
	sltBytes := []byte(salt)
	// Salt must be 8 bytes.
	if salt == "" {
		sltBytes = generateSalt()
		salt = string(sltBytes)
	}
	if len(sltBytes) != 8 {
		return "", fmt.Errorf("salt must be 8 bytes, given: %d", len(sltBytes))
	}

	digest := md5.New()
	digest.Write(pwBytes)
	digest.Write([]byte(Prefix))
	digest.Write(sltBytes)

	passwordLength := len(pwBytes)

	// we now add as many characters of the MD5(pw,salt,pw)
	altDigest := md5.New()
	altDigest.Write(pwBytes)
	altDigest.Write(sltBytes)
	altDigest.Write(pwBytes)
	alt := altDigest.Sum(nil)
	for ii := passwordLength; ii > 0; ii -= 16 {
		if ii > 16 {
			digest.Write(alt[:16])
		} else {
			digest.Write(alt[:ii])
		}
	}

	// This is a little odd, but is needed.
	buf := bytes.Buffer{}
	buf.Grow(passwordLength / 2)
	for ii := passwordLength; ii > 0; ii >>= 1 {
		if (ii & 1) == 1 {
			buf.WriteByte(0)
		} else {
			buf.WriteByte(pwBytes[0])
		}
	}

	digest.Write(buf.Bytes())
	buf.Reset()
	finalpw := digest.Sum(nil)

	// This is a weird concept, but here goes.
	// This essentially just re-hashes the password 1000 times,
	// but does so in various combinations of the password, and the salt.
	ctx := md5.New()
	for i := 0; i < Rounds; i++ {
		if (i & 1) == 1 {
			ctx.Write(pwBytes)
		} else {
			ctx.Write(finalpw[:16])
		}

		if i%3 != 0 {
			ctx.Write(sltBytes)
		}

		if i%7 != 0 {
			ctx.Write(pwBytes)
		}

		if (i & 1) == 1 {
			ctx.Write(finalpw[:16])
		} else {
			ctx.Write(pwBytes)
		}
		finalpw = ctx.Sum(nil)
		ctx.Reset()
	}

	// We're only going to read out 24 chars.
	buf.Grow(24)
	// 24 bits to base 64 for this
	fill := func(a byte, b byte, c byte) {
		v := uint(uint(c) | (uint(b) << 8) | (uint(a) << 16))
		for i := 0; i < 4; i++ { // and pump out a character for each 6 bits
			buf.WriteByte(validChars[v&0x3f])
			v >>= 6
		}
	}
	// The order of these indices is strange, be careful
	fill(finalpw[0], finalpw[6], finalpw[12])
	fill(finalpw[1], finalpw[7], finalpw[13])
	fill(finalpw[2], finalpw[8], finalpw[14])
	fill(finalpw[3], finalpw[9], finalpw[15])
	fill(finalpw[4], finalpw[10], finalpw[5])
	fill(0, 0, finalpw[11])

	// we then return the output string
	return Prefix + salt + "$" + string(buf.Bytes()[:22]), nil
}
