package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"

	"github.com/gofrs/uuid"
)

const (
	Numbers      = "0123456789"
	LowerChar    = "abcdefghijkmnpqrstuvwxyz"
	CapitalsChar = "ABCDEFGHJKMNPQRSTUVWXYZ"
)

// GenerateUUID 根据 uuidv4 生成 uuid
func GenerateUUID() (uuidString string, err error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return
	}
	uuidString = u4.String()
	return
}

// GenerateToken 生成 token
func GenerateToken(uuidString string) (token string, err error) {
	u1, err := uuid.NewV1()
	if err != nil {
		return
	}
	//先按 uuidv1 生成 uuid 字符串，再结合用户的 uuid 做哈希。将 token 的碰撞概率降到最低
	token = HashString(u1.String(), uuidString)
	return
}

//GenerateVcode 生成6位随机数字字符串
func GenerateVcode() (vcode string, err error) {
	result, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return
	}
	vcode = fmt.Sprintf("%06v", result)
	return
}

func GenerateRandString(number, lowerChar, capitalsChar bool, length int) (str string) {
	i := 0
	for {
		if number {
			nRand := mrand.Intn(len(Numbers))
			str = str + string(Numbers[nRand])
			i = i + 1
		}
		if i >= length {
			break
		}
		if lowerChar {
			lcRand := mrand.Intn(len(LowerChar))
			str = str + string(LowerChar[lcRand])
			i = i + 1
		}
		if i >= length {
			break
		}
		if capitalsChar {
			cpRand := mrand.Intn(len(CapitalsChar))
			str = str + string(CapitalsChar[cpRand])
			i = i + 1
		}
		if i >= length {
			break
		}
	}

	return
}
