// Package patch는 부분수정(PATCH) 요청용 3-state 필드 래퍼를 제공한다(모듈 공용).
// 순수 JSON 디코딩만 담당하며 DB 타입에 의존하지 않는다 — DB 매핑은 각 모듈에서
// 따로 한다(supervisor=pgtype / worker=database/sql Null*).
package patch

import "encoding/json"

// Field는 프론트의 {valid, value} 한 필드를 받는다. 3-state를 명시 페이로드로 표현한다:
//
//	필드 생략 또는 {valid:false}   → Set=false           (건드리지 않음)
//	{valid:true, value:null}      → Set=true, Value=nil  (컬럼을 NULL로)
//	{valid:true, value:x}         → Set=true, Value=&x   (값으로)
//
// 주의: valid("패치할지") ≠ DB의 non-null 여부. 둘은 변환부에서 분리한다.
type Field[T any] struct {
	Set   bool
	Value *T
}

func (f *Field[T]) UnmarshalJSON(b []byte) error {
	var raw struct {
		Valid bool            `json:"valid"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	f.Set = raw.Valid
	if !raw.Valid {
		return nil
	}
	if len(raw.Value) == 0 || string(raw.Value) == "null" {
		f.Value = nil // NULL로 세팅
		return nil
	}
	f.Value = new(T)
	return json.Unmarshal(raw.Value, f.Value)
}
