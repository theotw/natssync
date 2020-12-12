/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package bridgemodel

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/theotw/natssync/pkg/bridgemodel/errors"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"

	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewErrorResponse(subsystem, code string, message string, parameters map[string]string) *v1.ErrorResponse {
	if parameters == nil {
		parameters = make(map[string]string)
	}
	fullCode := fmt.Sprintf("%s.%s", subsystem, code)
	return &v1.ErrorResponse{
		Code:       fullCode,
		Message:    message,
		Parameters: parameters,
	}
}

func HandleErrors(ctx context.Context, errs ...error) (httpStatusCode int, resps []*v1.ErrorResponse) {
	for _, e := range errs {
		c, resp := HandleError(ctx, e)
		if c > httpStatusCode {
			httpStatusCode = c
		}
		resps = append(resps, resp)
	}
	return
}

var supportLocales = map[string]string{"en": "en", "en-US": "en"}

func getLocale(ctx context.Context) string {
	var ret string
	if c, ok := ctx.(*gin.Context); ok == true {
		header := c.GetHeader("Accept-Language")
		parts := strings.Split(header, ",")
		if len(parts) > 0 {
			cand := strings.TrimSpace(parts[0])
			x, ok := supportLocales[cand]
			if ok {
				ret = x
			}
		}
	}

	if len(ret) == 0 {
		ret = "en"
	}
	return ret
}
func HandleError(ctx context.Context, e error) (httpStatusCode int, resp *v1.ErrorResponse) {
	locale := getLocale(ctx)
	return HandleErrorWithLocale(locale, e)
}
func HandleErrorWithLocale(locale string, e error) (httpStatusCode int, resp *v1.ErrorResponse) {
	// Defaulting since there is currently no other supported languages

	switch e.(type) {
	case *errors.InternalError:
		x := e.(*errors.InternalError)
		log.Errorf("internal error %s", x.Error())
		for k, v := range x.Params {
			log.Errorf("error params %s = %v", k, v)
		}
		httpStatusCode = http.StatusBadRequest

		resp = NewErrorResponse(x.Subsystem, x.SubSystemError, errors.GetErrorString(locale, x.ErrorCode()), x.Params)

	default:
		params := make(map[string]string)
		if e != nil {
			log.Errorf("An unhandled error occurred: %T: %s", e, e.Error())
			params["detail"] = e.Error()
		} else {
			log.Errorf("Error is nil but in error handler")
		}
		httpStatusCode = http.StatusInternalServerError
		resp = NewErrorResponse(errors.BRIDGE_ERROR, errors.ERROR_CODE_UNKNOWN, errors.GetErrorString(locale, errors.ERROR_CODE_UNKNOWN), params)
	}
	return
}
