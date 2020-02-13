package user

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"oauth-server-lite/g"
)

func DeleteFailedCount(userID, clientIP string) {
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.Fail + userID + ":" + clientIP
	_, err := rc.Do("DEL", redisKey)
	if err != nil {
		log.Errorln(err)
	}
	return
}

func UpdateFailedCount(userID, clientIP string) (count int64) {
	rc := g.ConnectRedis().Get()
	defer rc.Close()
	redisKey := g.Config().RedisNamespace.Fail + userID + ":" + clientIP
	res, err := rc.Do("GET", redisKey)

	if err != nil {
		log.Errorln(err)
		return
	}
	if res == nil {
		count = 1
		_, err = rc.Do("SET", redisKey, count, "EX", g.Config().FailedIntervel)
		if err != nil {
			log.Errorln(err)
		}
		return
	}
	value := string(res.([]byte))
	ct, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Errorln(err)
	}
	count = ct + 1
	_, err = rc.Do("SET", redisKey, count, "EX", g.Config().FailedIntervel)
	if err != nil {
		log.Errorln(err)
	}
	return
}
