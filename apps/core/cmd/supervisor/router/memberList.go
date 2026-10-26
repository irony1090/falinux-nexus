package router

import (
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// 회원·장비 목록의 검색·필터·페이지 — 화-4(서버 목록), REF-member-ui.md.
// 접속 여부가 메모리(레지스트리)라 SQL로 못 거르므로 전부 읽어 Go에서 거른다(회원·장비 수백 개 규모)

const listMaxSize = 200

type listQuery struct {
	Q      string // 소문자
	Filter string
	Page   int // 1부터
	Size   int // 0 = 전부
}

type listResponse[T any] struct {
	Items  []T            `json:"items"`
	Total  int            `json:"total"`  // 필터 뒤 개수
	Page   int            `json:"page"`   // 범위 안으로 당긴 페이지
	Counts map[string]int `json:"counts"` // 검색 뒤 필터별 개수(알약 옆 숫자)
}

func parseListQuery(c echo.Context) listQuery {
	q := listQuery{
		Q:      strings.ToLower(strings.TrimSpace(c.QueryParam("q"))),
		Filter: c.QueryParam("f"),
		Page:   1,
	}
	if q.Filter == "" {
		q.Filter = "all"
	}
	if n, err := strconv.Atoi(c.QueryParam("page")); err == nil && n > 0 {
		q.Page = n
	}
	if n, err := strconv.Atoi(c.QueryParam("size")); err == nil && n > 0 {
		q.Size = min(n, listMaxSize)
	}
	return q
}

// pageList: 검색(match) -> 필터별 개수 -> 필터 -> 페이지. filters = 필터 이름 -> 판정("all"은 자동)
func pageList[T any](all []T, lq listQuery, match func(T) bool, filters map[string]func(T) bool) listResponse[T] {
	counts := map[string]int{"all": 0}
	for name := range filters {
		counts[name] = 0
	}
	picked := make([]T, 0, len(all))
	for _, it := range all {
		if lq.Q != "" && !match(it) {
			continue
		}
		counts["all"]++
		for name, f := range filters {
			if f(it) {
				counts[name]++
			}
		}
		if f, ok := filters[lq.Filter]; lq.Filter == "all" || (ok && f(it)) {
			picked = append(picked, it)
		}
	}
	out := listResponse[T]{Total: len(picked), Page: 1, Counts: counts, Items: picked}
	if lq.Size > 0 {
		pages := max(1, (len(picked)+lq.Size-1)/lq.Size)
		out.Page = min(lq.Page, pages)
		lo := (out.Page - 1) * lq.Size
		out.Items = picked[lo:min(lo+lq.Size, len(picked))]
	}
	return out
}

func containsFold(s, lowerQ string) bool {
	return strings.Contains(strings.ToLower(s), lowerQ)
}

// ===== 회원 =====

var userFilters = map[string]func(adminUserResponse) bool{
	"admin":     func(u adminUserResponse) bool { return slices.Contains(u.Perms, permAdmin) },
	"suspended": func(u adminUserResponse) bool { return u.SuspendedAt != nil },
	"noreg":     func(u adminUserResponse) bool { return !slices.Contains(u.Perms, permWorkerRegister) },
}

// ===== 장비 =====

func deviceState(d deviceResponse) string {
	if len(d.Instances) == 0 {
		return "none"
	}
	for _, i := range d.Instances {
		if i.Online {
			return "on"
		}
	}
	return "off"
}

var deviceFilters = map[string]func(deviceResponse) bool{
	"on":   func(d deviceResponse) bool { return deviceState(d) == "on" },
	"off":  func(d deviceResponse) bool { return deviceState(d) == "off" },
	"none": func(d deviceResponse) bool { return deviceState(d) == "none" },
}

// 장비 검색 = 이름·주인 닉네임·인스턴스 라벨
func deviceMatch(lowerQ string) func(deviceResponse) bool {
	return func(d deviceResponse) bool {
		if containsFold(d.MainKey, lowerQ) || containsFold(d.OwnerNickname, lowerQ) {
			return true
		}
		for _, i := range d.Instances {
			if containsFold(i.Label, lowerQ) {
				return true
			}
		}
		return false
	}
}
