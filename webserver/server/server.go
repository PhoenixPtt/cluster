package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"webserver/router"
)

// 供外部调用的初始化web服务端的方法
func Init() error {
	// 使用默认中间件创建一个gin路由器
	// logger and recovery (crash-free) 中间件
	serEngine := gin.Default()

	// 初始化路由对象，如果不能正常初始化，则返回false以及错误信息
	if !router.Init(serEngine) {
		err := fmt.Errorf("router initial is fail")
		return err
	} else { // 正常初始化，则执行可以实现优雅关闭web服务器的操作
		// 初始化一个web server
		srv := &http.Server{
			Addr:    ":8000",
			Handler: serEngine,
		}
		
		// Initializing the server in a goroutine so that
		// it won't block the graceful shutdown handling below
		go func() {
			log.Printf("cluster web server is listening %v\n", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("cluster web server is listen error: %s\n", err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
		quit := make(chan os.Signal)
		// kill (no param) default send syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down cluster web server...")

		// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("cluster web server forced to shutdown:", err)
		}

		log.Println("cluster web server exiting")

		return nil
	}
}
