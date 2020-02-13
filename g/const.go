package g

const (
	VERSION = "0.1.0"
	SALT    = "ba6bb05c50e03f6b5ab54a2b7914800d"
)

const (
	SUB                     = "sub"
	InvalidPassword         = "invalid_password"
	IpLocked                = "ip_locked"
	InvalidCaptcha          = "invalid_captcha"
	InvalidRequest          = "invalid_request"
	InvalidClient           = "invalid_client"
	InvalidIP               = "invalid_ip"
	InvalidGrant            = "invalid_grant"
	InvalidScope            = "invalid_scope"
	InvalidToken            = "invalid_token"
	InvaildRedirectUrl      = "invalid_redirect_url"
	UnauthorizedClient      = "unauthorized_client"
	UnsupportedGrantType    = "unsupported_grant_type"
	UnsupportedResponseType = "unsupported_response_type"
	AccessDenied            = "access_denied"
	ServerError             = "server_error"
)

var LoginErrorDescription = map[string]string{
	InvalidPassword: "用户名或密码不正确",
	ServerError:     "服务器内部错误",
	IpLocked:        "登录失败过多，ip 已经锁定",
	InvalidCaptcha:  "验证码不正确",
}

var OauthErrorDescription = map[string]string{
	InvalidRequest:          "请求缺少必需的参数、包含不支持的参数值（除了许可类型）、重复参数、包含多个凭据、采用超过一种客户端身份验证机制或其他不规范的格式",
	InvalidClient:           "客户端身份验证失败（例如，未知的客户端，不包含客户端身份验证，或不支持的身份验证方）",
	InvalidIP:               "请求的 IP 地址不在授权白名单内",
	InvaildRedirectUrl:      "请求的重定向 URL 格式不合法",
	InvalidGrant:            "提供的授权许可（如授权码、资源所有者凭据）或刷新令牌无效、过期、吊销、与在授权请求使用的重定向 URI 不匹配或颁发给另一个客户端",
	InvalidScope:            "请求的范围无效、未知的、格式不正确或超出资源所有者许可的范围",
	InvalidToken:            "令牌无效或者已经过期",
	UnauthorizedClient:      "进行身份验证的客户端没有被授权使用这种授权许可类型",
	UnsupportedGrantType:    "授权许可类型不被授权服务器支持",
	UnsupportedResponseType: "授权服务器不支持使用此方法获得授权码",
	AccessDenied:            "授权服务器拒绝该请求",
	ServerError:             "服务器内部错误",
}
