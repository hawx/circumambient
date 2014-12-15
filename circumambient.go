package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

type Json map[string]interface{}

var (
	in = flag.String("in", "", "")
	out = flag.String("out", "", "")
	console = flag.Bool("console", false, "")
	redisPort = flag.String("redis", ":6379", "")
	redisChannel = flag.String("channel", "requests", "")
	help = flag.Bool("help", false, "")
)

type Sender interface {
	Send(payload []byte)
}

type consoleSender struct {}

func NewConsoleSender() Sender {
	return &consoleSender{}
}

func (r *consoleSender) Send(payload []byte) {
	log.Println(string(payload))
}

type redisSender struct {
	pool    *redis.Pool
	channel string
}

func NewRedisSender(port, channel string) Sender {
	return &redisSender{
	  redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", port)
		}, 3),
    channel,
	}
}

func (r *redisSender) Send(payload []byte) {
	c := r.pool.Get()
	defer c.Close()

	err := c.Send("PUBLISH", r.channel, string(payload))
	if err != nil {
		log.Println(err)
	}
	c.Flush()
}

func main() {
	flag.Parse()

	if *help {
		fmt.Println(
			"Usage: circumambient [options]\n",
			"\n",
			"  Circumambient provides a simple proxy that publishes details of\n",
			"  the request to redis.\n",
			"\n",
			"    --in <host:port>      # What to route through the proxy\n",
			"    --out <host:port>     # Where to send proxied traffic\n",
			"    --redis <:port>       # Port of running redis server (default. :6379)\n",
			"    --channel <name>      # Name of channel to publish to (default: requests)\n",
		)

		os.Exit(0)
	}

	if *out == "" || *in == "" {
		fmt.Println("Usage: circumambient --in <host:port> --out <host:port>")
		os.Exit(2)
	}

	remote, err := url.Parse("http://" + *out)
	if err != nil {
		log.Fatal(err)
	}

	sender := NewConsoleSender()
	if !*console {
		sender = NewRedisSender(*redisPort, *redisChannel)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Println(r.Method + " " + r.URL.String())
		proxy.ServeHTTP(w, r)

		publish(sender, start, time.Since(start), r)
	})

	log.Println("started proxy on " + *in + " to " + *out + "...")
	log.Fatal(http.ListenAndServe(*in, nil))
}

func publish(sender Sender, start time.Time, duration time.Duration, r *http.Request) {
	headers := Json{}
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	payload, _ := json.Marshal(Json{
		"method":    r.Method,
		"url":       Json{
			"path":      r.URL.Path,
			"query":     r.URL.RawQuery,
			"fragment":  r.URL.Fragment,
		},
		"headers":   headers,
		"timestamp": start.UnixNano(),
		"duration":  duration.Nanoseconds(),
	})

	sender.Send(payload)
}
