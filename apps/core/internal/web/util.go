package web

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

type ClientError struct {
	Status int   `json:"status"`
	Error  error `json:"error"`
}

func Err(status int, err string, opt ...any) ClientError {
	return ClientError{
		Status: status,
		Error:  fmt.Errorf(err, opt...),
	}
}

// PanicMiddleware는 핸들러 내 panic을 recover하고 OnPanic을 호출함
func PanicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				// 필요하면 여기서 응답을 작성
				if !c.Response().Committed {
					if clientErr, ok := rec.(ClientError); ok {
						var typ string
						if clientErr.Status >= 500 {
							typ = "SERVER"
						} else {
							typ = "CLIENT"
						}
						_ = c.JSON(clientErr.Status, map[string]string{
							"message": clientErr.Error.Error(),
							"type":    typ,
						})
					} else {
						_ = c.JSON(500, map[string]string{
							"message": fmt.Sprintf("%v", rec),
							"type":    "SERVER",
						})
					}
				}
			}
		}()
		return next(c)
	}
}

func LogMiddelware(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {
		req := c.Request()
		loc := fmt.Sprintf("%s:%s", req.Method, req.URL.Path)
		// now := time.Now().Unix()
		remote := req.RemoteAddr
		log.Printf("[%s] (%s) START\n", loc, remote)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[%s] (%s) E N D - FAIL\n", loc, remote)
				panic(r)
			} else {
				log.Printf("[%s] (%s) E N D - SUCCESS\n", loc, remote)
			}
		}()
		return next(c)
	}
}
