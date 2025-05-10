package main

import (
	"blog/global"
	"blog/routers/router"
	"blog/setting"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	setting.Init()

	r := router.NewRouter()

	server := &http.Server{
		Addr:           global.Config.App.Port,
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server Name:", global.Config.App.Name)
	fmt.Println("server start at:", global.Config.App.Port)

	errChan := make(chan error, 1)
	defer close(errChan)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		fmt.Println("err:", err)
	case <-quit:
		fmt.Println("server shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Println("Server Shutdown:", err)
		}
	}
}
