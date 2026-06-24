package util

import (
	"crypto/rand"
	"fmt"
)

// base62: 숫자 + 영문 대소문자 (62글자)
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// 256을 62로 나눈 가장 큰 배수(=248). 이 값 이상의 바이트는 버려 modulo 편향을 없앤다.
const maxByte = byte(256 - 256%len(alphabet))

// RandomKey는 prefix + 랜덤(base62) + suffix 형태의 키를 만든다.
// length는 prefix/suffix까지 포함한 전체 길이이며, 결과는 정확히 length 글자다.
// prefix+suffix가 length보다 길면 에러를 반환한다.
func RandomKey(length int, prefix, suffix string) (string, error) {
	fixed := len(prefix) + len(suffix)
	if length < fixed {
		return "", fmt.Errorf("util: length(%d)가 prefix+suffix(%d)보다 작음", length, fixed)
	}
	n := length - fixed // 랜덤으로 채울 글자 수

	out := make([]byte, 0, n)
	for len(out) < n {
		buf := make([]byte, n-len(out)) // 부족한 만큼만 더 읽음
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if b >= maxByte {
				continue // 편향 구간(248~255)은 버림
			}
			out = append(out, alphabet[b%byte(len(alphabet))])
			if len(out) == n {
				break
			}
		}
	}
	return prefix + string(out) + suffix, nil
}
