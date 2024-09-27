package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"oauth-server-lite/controller"
	"oauth-server-lite/g"
	"oauth-server-lite/models/oauth"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)
	srv := controller.InitGin(g.Config().Http.Listen)
	g.InitLog(g.Config().Logger)

	g.InitDB()
	oauth.InitTables()
	g.InitRedisConnPool()

	defer g.CloseRedis()
	defer g.CloseDB()

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
		log.Fatal("Server Shutdown: %s", err)
	}
	log.Println("Server exit")
}
