package main

import (
	"flag"
	"log"
	"net/http"

	"cafeteria-ordering/internal/cafeteria"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP 服务监听地址")
	flag.Parse()
	log.SetFlags(0)

	store := cafeteria.NewStore()
	server := cafeteria.NewHTTPServer(store)

	log.Printf("食堂订餐服务已启动: http://%s", *addr)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
