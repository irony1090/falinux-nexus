package protocol

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeSupervisorAddr는 supervisor 주소를 "ws://host:port" 꼴로 맞춘다(경로 /worker는 worker가 붙임) —
// supervisor(명령 받을 때)·worker(저장할 때) 같은 검사, 주소-2(옮기기 명령)
func NormalizeSupervisorAddr(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("주소 해석 실패: %w", err)
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return "", fmt.Errorf("ws 또는 wss 주소여야 합니다: %q", raw)
	}
	if u.Host == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.User != nil {
		return "", fmt.Errorf("scheme://host 꼴이어야 합니다: %q", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}
