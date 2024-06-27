package midd

import (
	"net/http"
	location_utils "oauth-server-lite/controller/location-utils"

	"github.com/gin-gonic/gin"
	"oauth-server-lite/g"
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

type OauthAppApi struct {
	AppId          int64  `gorm:"not null;default:0;comment:应用ID"`
	AppName        string `gorm:"not null;default:'';comment:应用名称"`
	Description    string `gorm:"default:'';comment:应用描述"`
	ApiId          string `gorm:"not null;default:'';comment:应用对应授权的APIID"`
	ApiName        string `gorm:"not null;default:'';comment:应用对应授权的API名称"`
	ApiDescription string `gorm:"default:'';comment:应用对应授权的API描述"`
}

/*
AuthorizeHTML 授权页渲染
*/
func AuthorizeHTML(userId, appName, privacyUrl string, apis []OauthAppApi, c *gin.Context, name string) {
	location := location_utils.GetLocation(c.Request)
	contextPath := location.Scheme + "://" + location.Host

	c.HTML(http.StatusOK, name, gin.H{
		"contextPath": contextPath,
		"userId":      userId,
		"appName":     appName,
		"privacyUrl":  privacyUrl,
		"authApis":    apis,
	})
}

/*
LogoutHTML 登录页渲染
*/
func LogoutHTML(c *gin.Context) {
	location := location_utils.GetLocation(c.Request)
	contextPath := location.Scheme + "://" + location.Host

	c.HTML(http.StatusOK, "logout.html", gin.H{
		"contextPath": contextPath,
	})
}

/*
ErrorHTML 错误页渲染
*/
func ErrorHTML(errorMsg string, c *gin.Context) {
	c.HTML(http.StatusOK, "p2.html", gin.H{
		"errorMsg": errorMsg,
	})
}

// APIResult api 接口的数据结构
type APIResult struct {
	ErrCode int64       `json:"errCode"`
	ErrMsg  string      `json:"errMsg"`
	Data    interface{} `json:"data"`
}

const (
	Success                         = 0
	LdapInvalidPasswordOrOtherError = 10001
	LdapInvalidUsernameOrLocked     = 10002
	ParamFormatError                = 1108110301
	ParamValueError                 = 1108110304
	ParamMissError                  = 1108110519
	InternalAPIError                = 1108110525
	WxWorkUserNotFound              = "WxWorkUserNotFound"
	UserNotFound                    = "UserNotFound"
)

var codeMsg = map[int64]string{
	Success:                         "success",
	ParamFormatError:                "参数校验错误",
	ParamValueError:                 "参数取值错误",
	ParamMissError:                  "缺失参数",
	InternalAPIError:                "内部API错误",
	LdapInvalidPasswordOrOtherError: "不正确的LDAP密码",
	LdapInvalidUsernameOrLocked:     "用户不存在或已被禁用",
}

// ErrorRes 请求异常时的返回
func ErrorRes(code int64, msg string) (res APIResult) {
	res.ErrCode = code
	if msg == "" {
		res.ErrMsg = codeMsg[code]
	} else {
		res.ErrMsg = msg
	}
	return
}

// SuccessRes 请求正常时的返回
func SuccessRes(data interface{}) (apiResult APIResult) {
	apiResult.ErrCode = 0
	apiResult.ErrMsg = "success"
	apiResult.Data = data
	return
}
