package midd

import (
	"net/http"

	"oauth-server-lite/g"

	"github.com/gin-gonic/gin"
)

type OauthErrorResult struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func OauthErrorRes(err string) OauthErrorResult {
	oauthErrorRes := OauthErrorResult{
		Error:            err,
		ErrorDescription: g.OauthErrorDescription[err],
	}
	return oauthErrorRes

}

/*
LoginHTML 登录页渲染
*/
func LoginHTML(errorMsg, captchaId string, c *gin.Context) {
	location := GetLocation(c.Request)
	contextPath := location.Scheme + "://" + location.Host

	c.HTML(http.StatusOK, "index.html", gin.H{
		"contextPath": contextPath,
		"error_msg":   errorMsg,
		"captcha_id":  captchaId,
	})
}
