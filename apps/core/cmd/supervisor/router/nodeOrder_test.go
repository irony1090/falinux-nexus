package router

import (
	"slices"
	"testing"
)

func TestPlaceSiblings(t *testing.T) {
	p := func(v int64) *int64 { return &v }
	sibs := []int64{1, 2, 3, 4, 5}
	cases := []struct {
		name   string
		moving []int64
		after  *int64
		want   []int64
	}{
		{"뒤로", []int64{2}, p(4), []int64{1, 3, 4, 2, 5}},
		{"앞으로", []int64{5}, p(1), []int64{1, 5, 2, 3, 4}},
		{"맨 앞", []int64{3}, nil, []int64{3, 1, 2, 4, 5}},
		{"없는 after = 맨 뒤", []int64{1}, p(99), []int64{2, 3, 4, 5, 1}},
		{"여럿은 형제 순서", []int64{4, 1}, p(5), []int64{2, 3, 5, 1, 4}},
		{"제자리", []int64{3}, p(2), []int64{1, 2, 3, 4, 5}},
	}
	for _, c := range cases {
		got, bad := placeSiblings(sibs, c.moving, c.after)
		if bad != "" || !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v %q, want %v", c.name, got, bad, c.want)
		}
	}
	if _, bad := placeSiblings(sibs, []int64{9}, nil); bad == "" {
		t.Error("형제 아닌 노드를 받아들임")
	}
	if _, bad := placeSiblings(sibs, []int64{2, 3}, p(3)); bad == "" {
		t.Error("옮기는 노드 뒤로 놓기를 받아들임")
	}
}
