package user

import (
	log "github.com/sirupsen/logrus"

	"oauth-server-lite/g"
)

func CreateLock(userID, clientIP string) {
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.Lock + userID + ":" + clientIP
	_, err := rc.Do("SET", redisKey, 1, "EX", g.Config().LockTime)
	if err != nil {
		log.Errorln(err)
	}
	return
}

func CheckLock(userID, clientIP string) (lock bool) {
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.Lock + userID + ":" + clientIP
	res, err := rc.Do("GET", redisKey)

	if err != nil {
		log.Errorln(err)
		return
	}
	if res != nil {
		lock = true
		return
	}
	return
}
