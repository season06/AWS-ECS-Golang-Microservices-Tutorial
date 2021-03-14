package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

var RedisEndpoint = "localhost:6379" // two container in a same task definition (same ENI)

var CLIENT = RedisClient()

type Content struct {
	IP    string
	Count int
}

func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     RedisEndpoint,
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println(pong, err)
	} else {
		fmt.Println("Connect Redis Success!")
	}

	return client
}

func countIP(w http.ResponseWriter, r *http.Request) {
	client_ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	var counting int

	counting, err := CLIENT.Get(client_ip).Int()
	if err == redis.Nil {
		counting = 1
		CLIENT.Set(client_ip, counting, 60*time.Second)
	} else {
		ttl, _ := CLIENT.TTL(client_ip).Result()
		counting++
		CLIENT.Set(client_ip, counting, ttl)

		if ttl < time.Duration(0) {
			CLIENT.Del(client_ip)
		}
	}

	// show the result in HTML
	tmpl := template.Must(template.ParseFiles("./index.html"))

	data := Content{
		IP:    client_ip,
		Count: counting,
	}
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/home", countIP)

	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}
