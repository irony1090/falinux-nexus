package util

func IsCompletedUTF8(data []byte) (int, int) {
	if len(data) == 0 {
		return -1, 0
	}

	// 끝에서부터 역순으로 확인
	lastIdx := len(data) - 1

	// fmt.Printf("[IsCompletedUTF8] %b\n", data)
	// fmt.Printf("[IsCompletedUTF8] %b & %b -> %b\n", data[lastIdx], 0x80, data[lastIdx]&0x80)
	if data[lastIdx]&0x80 == 0 {
		return lastIdx, 0
	}
	// continuation byte (10xxxxxx)를 역순으로 찾기
	continuationBytes := 0
	for i := lastIdx; i >= 0 && i > lastIdx-4; i-- {
		b := data[i]
		if b&0xc0 == 0x80 {
			continuationBytes++
			continue
		}

		// 시작 바이트 발견
		expectedBytes := 0
		if b&0xE0 == 0xC0 { // 110xxxxx (2바이트 문자)
			expectedBytes = 2
		} else if b&0xF0 == 0xE0 { // 1110xxxx (3바이트 문자)
			expectedBytes = 3
		} else if b&0xF8 == 0xF0 { // 11110xxx (4바이트 문자)
			expectedBytes = 4
		}

		actualBytes := continuationBytes + 1
		if actualBytes == expectedBytes {
			return lastIdx, 0 // 완성형
		} else if actualBytes < expectedBytes {
			return -1, expectedBytes - actualBytes // 불완전, 부족한 바이트 수 반환
		} else {
			return -1, -1 // 잘못된 형식
		}
	}

	return -1, -1
}
