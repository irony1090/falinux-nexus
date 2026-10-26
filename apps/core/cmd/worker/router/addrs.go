package router

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"time"

	"nexus/internal/protocol"
	workerdb "nexus/internal/worker/db/gen"
	"nexus/internal/worker/store"
)

// ===== supervisor 주소 목록 — 주소-1(주소 목록), REF-supervisor-move.md =====
// 순위대로 접속 시도, 붙은 주소 = 1순위. 옮기기 명령의 새 주소도 1순위로 들어온다(주소-Q10(새 주소 순위))

const MaxAddrs = 3 // 주소-Q13(N 값)

const settingEnvSeed = "env_seed"

type AddrBook struct {
	mu    sync.Mutex
	store *store.StorePool
	list  []string
}

// WorkerURL은 목록 주소에 worker 접속 경로를 붙인다
func WorkerURL(addr string) url.URL {
	u, _ := url.Parse(addr)
	u.Path = "/worker"
	return *u
}

// LoadAddrBook은 저장된 목록을 읽고 env 주소(seed)를 반영한다. env 값이 지난번과 다를 때만(사용자가 env를 고침) 1순위로 —
// 같으면 옮기기 명령이 정한 순위를 지킨다(재시작마다 env 주소가 1순위로 돌아오면 옮긴 게 풀림)
func LoadAddrBook(st *store.StorePool, seed string) (*AddrBook, error) {
	seed, err := protocol.NormalizeSupervisorAddr(seed)
	if err != nil {
		return nil, fmt.Errorf("env WS_SCHEME·WS_HOST: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := st.Queries()
	list, err := q.ListSupervisors(ctx)
	if err != nil {
		return nil, err
	}
	last, err := q.GetSetting(ctx, settingEnvSeed)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	b := &AddrBook{store: st, list: list}
	if len(list) == 0 || last != seed {
		if _, err := b.Promote(seed); err != nil {
			return nil, err
		}
		if err := q.SetSetting(ctx, workerdb.SetSettingParams{Key: settingEnvSeed, Value: seed}); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (b *AddrBook) List() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.list)
}

// Promote는 addr을 1순위로 올리고(없으면 넣음) MaxAddrs를 넘는 꼴찌를 지운 뒤 저장한다. 저장 뒤 목록을 돌려준다
func (b *AddrBook) Promote(addr string) ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	next := append([]string{addr}, slices.DeleteFunc(slices.Clone(b.list), func(a string) bool { return a == addr })...)
	if len(next) > MaxAddrs {
		next = next[:MaxAddrs]
	}
	if slices.Equal(next, b.list) {
		return slices.Clone(next), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.store.Tx(ctx, func(q *workerdb.Queries) error {
		if err := q.DeleteSupervisors(ctx); err != nil {
			return err
		}
		for i, a := range next {
			if err := q.InsertSupervisor(ctx, workerdb.InsertSupervisorParams{Url: a, Rank: int64(i)}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("주소 목록 저장 실패: %w", err)
	}
	b.list = next
	return slices.Clone(next), nil
}
