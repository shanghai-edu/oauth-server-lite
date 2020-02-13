package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/satori/go.uuid"
)

// HashString 按 sha256 加盐生成 hash 字符串
func HashString(s, salt string) (hashedString string) {
	hash := sha256.New()
	hash.Write([]byte(s + salt))
	hashedString = hex.EncodeToString(hash.Sum(nil))
	return
}

// RandHashString 生成一个随机的 hash 字符串，保留 x 位
func RandHashString(salt string, l int) (hashedString string, err error) {
	if l <= 0 {
		err = errors.New("lengh is must bigger than 0")
		return
	}

	u1 := uuid.NewV4()

	//先按 uuidv4 生成随机的 uuid 字符串，再加盐哈希
	str := HashString(u1.String(), salt)
	if l > len(str) {
		err = errors.New("lengh is too long")
		return
	}
	hashedString = str[0:l]
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
