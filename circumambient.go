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
	pool *redis.Pool
	in = flag.String("in", "", "")
	out = flag.String("out", "", "")
	redisPort = flag.String("redis", ":6379", "")
	redisChannel = flag.String("channel", "requests", "")
	help = flag.Bool("help", false, "")
)

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

	pool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", *redisPort)
	}, 3)

	proxy := httputil.NewSingleHostReverseProxy(remote)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Println(r.Method + " " + r.URL.String())
		proxy.ServeHTTP(w, r)

		publish(start, time.Since(start), r)
	})

	log.Println("started proxy on " + *in + " to " + *out + "...")
	log.Fatal(http.ListenAndServe(*in, nil))
}

func publish(start time.Time, duration time.Duration, r *http.Request) {
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

	c := pool.Get()
	defer c.Close()

	err := c.Send("PUBLISH", *redisChannel, string(payload))
	if err != nil {
		log.Println(err)
	}
	c.Flush()
}
