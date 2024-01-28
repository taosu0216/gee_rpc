package main

import (
	"fmt"
	"gee_RPC/client"
	"gee_RPC/server"
	"log"
	"net"
	"sync"
	"time"
)

func start(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	server.Accept(l)
}
func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go start(addr)
	conn, _ := client.Dail("tcp", <-addr)
	defer func() {
		_ = conn.Close()
	}()
	time.Sleep(time.Second)
	//_ = json.NewEncoder(conn).Encode(server.DefaultOption)
	//cc := codec.NewGobCodec(conn)
	//for i := 0; i < 5; i++ {
	//	h := &codec.Header{
	//		ServiceMethod: "Foo.Sum",
	//		Seq:           uint(uint64(i)),
	//	}
	//	_ = cc.Write(h, fmt.Sprintf("gee rpc req %d!", h.Seq))
	//	_ = cc.ReadHeader(h)
	//	var reply string
	//	_ = cc.ReadBody(&reply)
	//	log.Println("reply:", reply)
	//}
	//     day2
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("gee rpc req %d!", i)
			fmt.Println()
			var reply string
			if err := conn.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
		wg.Wait()
	}
}
