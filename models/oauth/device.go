package oauth

import (
	"encoding/json"
	"errors"
	"oauth-server-lite/controller/location-utils"
	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/logger"
)

type DeviceCodeInput struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	UserID   string `json:"user_id"`
}

type DeviceCodeOutput struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiresIn       int64  `json:"expires_in"`
	Interval        int64  `json:"interval"`
}

type DeviceTokenInput struct {
	GrantType  string
	ClientID   string
	DeviceCode string
}

func CreateDeviceCode(c *gin.Context, inputs DeviceCodeInput) (deviceCodeOutput DeviceCodeOutput, err error) {
	userCode, err1 := utils.RandHashString(g.SALT, 16)
	deviceCode, err2 := utils.RandHashString(g.SALT, 32)
	if err1 != nil || err2 != nil {
		return
	}

	location := location_utils.GetLocation(c.Request)
	verificationUri := location.Scheme + "://" + location.Host + "/user/device/authorize"
	deviceCodeOutput = DeviceCodeOutput{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationUri: verificationUri,
		ExpiresIn:       g.Config().CodeExpired,
		Interval:        15,
	}
	input, err := json.Marshal(inputs)
	output, err := json.Marshal(deviceCodeOutput)
	if err != nil {
		return
	}
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	// 存入device_code对应的input和output
	userCodeKey := g.Config().RedisNamespace.OAuth + "device_code_output:" + userCode
	deviceKey := g.Config().RedisNamespace.OAuth + "device_code_input:" + deviceCode
	_, err = rc.Do("SET", userCodeKey, string(output), "EX", g.Config().CodeExpired)
	_, err = rc.Do("SET", deviceKey, string(input), "EX", g.Config().CodeExpired)
	return
}

func CheckDeviceCode(inputs DeviceTokenInput) (err error) {
	logger.Debugf("check device_code: %v", inputs)
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	deviceKey := g.Config().RedisNamespace.OAuth + "device_code_input:" + inputs.DeviceCode
	redisDeviceCodeInput, err := rc.Do("GET", deviceKey)

	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}
	if redisDeviceCodeInput == nil {
		err = errors.New(g.InvalidGrant)
		return
	}
	// 校验client_id是否一致
	var device DeviceCodeInput
	err = json.Unmarshal(redisDeviceCodeInput.([]byte), &device)
	if err != nil {
		logger.Error(err)
		err = errors.New(g.ServerError)
		return
	}

	if !(inputs.ClientID == device.ClientID) {
		err = errors.New(g.InvalidGrant)
		return
	}

	return
}
