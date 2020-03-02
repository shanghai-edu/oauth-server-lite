package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toolkits/file"

	"oauth-server-lite/controller"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	init := flag.Bool("i", false, "init db tables")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)
	g.InitLog(g.Config().LogLevel)

	if *init {
		if err := oauth.InitTables(); err != nil {
			log.Fatalf("Init DB Failed: %s", err)
		}
		os.Exit(0)
	}

	if g.Config().DB.Sqlite != "" {
		if file.IsExist(g.Config().DB.Sqlite) {
			if !file.IsFile(g.Config().DB.Sqlite) {
				log.Fatalf(g.Config().DB.Sqlite + "is not directory, not file")
			}
		} else {
			if err := oauth.InitTables(); err != nil {
				log.Fatalf("Init DB Failed: %s", err)
			}
		}
	}

	err := g.InitDB(g.Config().DB.DBDebug)
	if err != nil {
		log.Fatalf("db conn failed with error %s", err.Error())
	}
	defer g.CloseDB()

	g.InitRedisConnPool()
	defer g.CloseRedis()

	srv := controller.InitGin(g.Config().Http.Listen)

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %s", err)
	}
	log.Println("Server exit")
}
