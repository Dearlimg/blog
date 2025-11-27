package main

import (
	"blog/global"
	"blog/pkg/logger"
	"blog/routers/router"
	"blog/setting"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	setting.Init()
	defer logger.Sync()

	r := router.NewRouter()

	server := &http.Server{
		Addr:           global.Config.App.Port,
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("Server starting",
		logger.String("name", global.Config.App.Name),
		logger.String("port", global.Config.App.Port),
		logger.String("version", global.Config.App.Version),
	)

	errChan := make(chan error, 1)
	defer close(errChan)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", logger.ErrorField(err))
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.Error("Server error", logger.ErrorField(err))
	case <-quit:
		logger.Info("Server shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", logger.ErrorField(err))
		} else {
			logger.Info("Server shutdown successfully")
		}
	}
}
