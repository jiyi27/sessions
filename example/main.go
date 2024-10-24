package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/shwezhu/sessions"
)

func DoNothing(_ http.ResponseWriter, _ *http.Request) {}

func handler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-id")

	if session.IsNew() {
		session.SetValue("name", "Coco")
		session.SetValue("age", 18)
		// Set max age for session in seconds.
		session.SetMaxAge(30)
		// Save session into response to client.
		// You should save session after you make change on a session.
		session.Save(w)
		return
	}

	name := session.GetValueByKey("name")
	_, _ = fmt.Fprint(w, fmt.Sprintf("hello %v\n", name))
}

var store *sessions.RedisStore

func main() {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果没有密码，置空
		DB:       0,  // 使用默认DB
	})
	store = sessions.NewRedisStore(rdb)

	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", DoNothing)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
