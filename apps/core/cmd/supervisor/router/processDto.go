package router

import (
	superdb "nexus/internal/supervisor/db/gen"
)

// process 응답 DTO + DTO↔pgtype 변환. 핸들러는 processApi.go. nodeDto.go와 대칭.

type processResponse struct {
	Uid        string   `json:"uid"`
	Type       string   `json:"type"`
	NodeID     *int64   `json:"nodeId"`
	DeviceKey  string   `json:"deviceKey"`
	Cmd        string   `json:"cmd"`
	Args       []string `json:"args"`
	Env        []string `json:"env"`
	Cwd        string   `json:"cwd"`
	Rows       int16    `json:"rows"`
	Cols       int16    `json:"cols"`
	Status     string   `json:"status"`
	Pid        *int32   `json:"pid"`
	ExitCode   *int32   `json:"exitCode"`
	CreatedAt  int64    `json:"createdAt"`
	StartedAt  *int64   `json:"startedAt"`
	FinishedAt *int64   `json:"finishedAt"`
	UpdatedAt  *int64   `json:"updatedAt"`
	SizeOwner  *bool    `json:"sizeOwner,omitempty"` // 요청 탭 기준이라 REST 응답에만 싣는다(토픽 발행 땐 nil)
}

func newProcessResponse(p superdb.Process) processResponse {
	r := processResponse{
		Uid: p.Uid, Type: p.Type, DeviceKey: p.DeviceKey,
		Cmd: p.Cmd, Args: p.Args, Env: p.Env, Cwd: p.Cwd,
		Rows: p.Rows, Cols: p.Cols, Status: p.Status,
	}
	if p.NodeID.Valid {
		v := p.NodeID.Int64
		r.NodeID = &v
	}
	if p.Pid.Valid {
		v := p.Pid.Int32
		r.Pid = &v
	}
	if p.ExitCode.Valid {
		v := p.ExitCode.Int32
		r.ExitCode = &v
	}
	if p.CreatedAt.Valid {
		r.CreatedAt = p.CreatedAt.Time.Unix()
	}
	if p.StartedAt.Valid {
		v := p.StartedAt.Time.Unix()
		r.StartedAt = &v
	}
	if p.FinishedAt.Valid {
		v := p.FinishedAt.Time.Unix()
		r.FinishedAt = &v
	}
	if p.UpdatedAt.Valid {
		v := p.UpdatedAt.Time.Unix()
		r.UpdatedAt = &v
	}
	return r
}

func withSizeOwner(r processResponse, owner bool) processResponse {
	r.SizeOwner = &owner
	return r
}

func newProcessResponses(ps []superdb.Process) []processResponse {
	out := make([]processResponse, len(ps))
	for i, p := range ps {
		out[i] = newProcessResponse(p)
	}
	return out
}
