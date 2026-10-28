package router

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"strconv"
	"strings"

	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 보기 타일 화면 표시 — 파트(보기 타일), REF-file-view.md. 내려받기와 달리 등록 없이 GET 한 번(<img src>에 그대로)

const probeSize = 8 * 1024 // 바이너리 판정에 읽는 앞부분(git과 같은 NUL 검사)

// viewTypes: V1(확장자) — 프론트 fileKind.util.ts와 같은 목록
var viewTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".svg":  "image/svg+xml",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
}

// fileExt: 소문자 확장자. Windows 경로(`\`)도 마지막 이름에서만 찾는다(filepath.Ext는 supervisor OS 기준)
func fileExt(path string) string {
	name := path[strings.LastIndexAny(path, `/\`)+1:]
	i := strings.LastIndex(name, ".")
	if i <= 0 {
		return ""
	}
	return strings.ToLower(name[i:])
}

// fileQuery: ?folderId(연 폴더) &path(파일 절대 경로) 검사 + 인가(DB). online = 지금 접속 중(아니면 보기는 사본으로)
func (r *supervisorRouter) fileQuery(c echo.Context) (userID int64, authKey, path string, online bool) {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	path = c.QueryParam("path")
	if path == "" {
		panic(web.Err(400, "path가 없습니다"))
	}
	requireDeviceFolder(c, sess.Data.ID, folderID, mainKey)
	EndTx(c)
	authKey = mainKey + "#" + subKey
	_, online = r.workers.Get(authKey)
	return sess.Data.ID, authKey, path, online
}

// viewHeaders: 공통 헤더. true = If-None-Match가 같아 304로 끝냄(사본·worker 둘 다 ETag = 크기+수정 시각)
func viewHeaders(c echo.Context, ctype, path string, m viewMeta) bool {
	h := c.Response().Header()
	etag := fmt.Sprintf(`"%x-%x"`, m.Size, m.ModTime)
	h.Set("ETag", etag)
	h.Set("Cache-Control", "private, no-cache")
	if c.Request().Header.Get("If-None-Match") == etag {
		c.NoContent(304)
		return true
	}
	h.Set(echo.HeaderContentType, ctype)
	h.Set(echo.HeaderContentDisposition, mime.FormatMediaType("inline", map[string]string{"filename": fileBase(path)}))
	h.Set("Content-Security-Policy", "sandbox") // V2(SVG 보안): SVG 안 스크립트가 Nexus 주소 권한으로 돌지 않게
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set(echo.HeaderContentLength, strconv.FormatInt(m.Size, 10))
	return false
}

// GET /workers/:mainKey/:subKey/file: 화면 표시용 본문.
// 접속 중 = 크기·수정 시각이 사본과 같으면 사본, 다르면 worker에서 받으며 사본을 새로 씀 / 끊김 = 사본(없으면 409)
func (r *supervisorRouter) viewWorkerFile(c echo.Context) error {
	userID, authKey, path, online := r.fileQuery(c)
	ctype, ok := viewTypes[fileExt(path)]
	if !ok {
		panic(web.Err(415, "보기를 지원하지 않는 형식입니다"))
	}
	if !online {
		f, m, ok := r.view.get(userID, authKey, path)
		if !ok {
			panic(web.Err(409, "접속 중인 인스턴스가 아닙니다"))
		}
		defer f.Close()
		return serveViewCopy(c, ctype, path, m, f)
	}

	src, err := r.openWorkerFile(c.Request().Context(), authKey, path, nil)
	if err != nil {
		panicOpenError(err)
	}
	defer src.Close()
	m := viewMeta{Size: src.Size(), ModTime: src.ModTime()}
	if f, cm, ok := r.view.get(userID, authKey, path); ok {
		defer f.Close()
		if cm == m {
			return serveViewCopy(c, ctype, path, m, f)
		}
	}

	if viewHeaders(c, ctype, path, m) {
		return nil
	}
	w := r.view.create(userID, authKey, path, m)
	c.Response().WriteHeader(200)
	var done int64
	for {
		data, err := src.Read(sendChunkSize)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			w.abort()
			abortConn(c)
			return nil
		}
		w = w.write(data)
		if _, err := c.Response().Write(data); err != nil {
			w.abort()
			return nil
		}
		done += int64(len(data))
	}
	if done != m.Size {
		w.abort()
		abortConn(c) // 읽는 중 파일이 바뀜 — 잘린 이미지를 정상으로 보이지 않게
		return nil
	}
	w.commit()
	return nil
}

func serveViewCopy(c echo.Context, ctype, path string, m viewMeta, f *os.File) error {
	if viewHeaders(c, ctype, path, m) {
		return nil
	}
	c.Response().WriteHeader(200)
	if _, err := io.Copy(c.Response(), f); err != nil {
		abortConn(c)
	}
	return nil
}

type workerFileProbeResponse struct {
	Binary bool `json:"binary"`
}

// GET /workers/:mainKey/:subKey/file/probe: 앞 probeSize에 NUL이 있으면 바이너리 — D(더블클릭 열기)에서 편집기로 열지 고름
func (r *supervisorRouter) probeWorkerFile(c echo.Context) error {
	_, authKey, path, online := r.fileQuery(c)
	if !online {
		panic(web.Err(409, "접속 중인 인스턴스가 아닙니다"))
	}
	src, err := r.openWorkerFile(c.Request().Context(), authKey, path, nil)
	if err != nil {
		panicOpenError(err)
	}
	defer src.Close()
	data, err := src.Read(probeSize)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(web.Err(502, "%v", err))
	}
	return c.JSON(200, workerFileProbeResponse{Binary: bytes.IndexByte(data, 0) >= 0})
}
