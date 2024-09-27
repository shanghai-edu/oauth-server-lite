package oauth

import (
	//	"encoding/json"
	"log"
	"oauth-server-lite/g"
)

func init() {
	g.ParseConfig("cfg.json")
	err := g.InitDB()
	if err != nil {
		log.Fatalf("db conn failed with error %s", err.Error())
	}
	g.InitRedisConnPool()
}

/*
func Test_InitTables(t *testing.T) {
	err := InitTables()
	if err != nil {
		t.Error(err)
	}
}
*/
