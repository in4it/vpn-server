package totp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const INTERVAL = 30

func GetToken(secret string, interval int64) (string, error) {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("base32 decode error: %s", err)
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(interval))
	hmacHash := hmac.New(sha1.New, key)
	hmacHash.Write(buf)
	h := hmacHash.Sum(nil)
	offset := (h[19] & 15)

	var header uint32
	r := bytes.NewReader(h[offset : offset+4])
	err = binary.Read(r, binary.BigEndian, &header)

	if err != nil {
		return "", fmt.Errorf("binary read error: %s", err)
	}

	return fmt.Sprintf("%06d", int((int(header)&0x7fffffff)%1000000)), nil
}

func Verify(secret, code string) (bool, error) {
	token, err := GetToken(secret, time.Now().Unix()/30)
	if err != nil {
		return false, fmt.Errorf("GetToken error: %s", err)
	}
	return token == code, nil
}

func VerifyMultipleIntervals(secret, code string, count int) (bool, error) {
	return verifyMultipleIntervals(secret, code, count, time.Now())
}

func verifyMultipleIntervals(secret, code string, count int, now time.Time) (bool, error) {
	for i := 0; i < count; i++ {
		token, err := GetToken(secret, now.Add(time.Duration(i)*time.Duration(-30)*time.Second).Unix()/30)
		if err != nil {
			return false, fmt.Errorf("GetToken error: %s", err)
		}
		if token == code {
			return true, nil
		}
	}
	return false, nil
}
