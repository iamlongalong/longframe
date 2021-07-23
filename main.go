package main

import (
	"night-fury/dashboard"
	"night-fury/pkgs/log"
	"night-fury/pkgs/utils"

	"net/http"
	_ "net/http/pprof"
	_ "night-fury/docs"

	_ "gitlab.lanhuapp.com/gopkgs/config"
)

func init() {
	// 设置最大进程数
	utils.SetMaxProcs()

	// 开启 pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
}

func main() {
	apiServer := dashboard.NewServer()

	go apiServer.Serve()

	sig, err := utils.GraceShutdown([]func() error{
		func() error { // 退出回调函数，例如 kafka 断联等
			return nil
		},
	})

	log.Infof(log.TagInit, "server shutdown via signal: %v, err : %s", sig, err)
}
