# Circumambient

Simple proxy to log details of any request made through it to redis.


## Getting started

Install [redis](http://redis.io/topics/quickstart) and start `redis-server` if
it isn't already running.

Then put this repo somewhere and run the proxy,

``` bash
$ git clone https://github.com/hawx/circumambient.git
$ cd circumambient
$ go build
$ ./circumambient --in localhost:3002 --out localhost:3001
```

Now any requests to <localhost:3002> will be passed to <localhost:3001>, but any
redis subscribers listening on the channel "requests" will get the following
message:

``` json
{
  "method": "...",
  "url": "...",
  "headers": {
    // ...
  },
  "timestamp": 1397731289, // unix timestamp when request was made
  "duration": 143314       // duration of request in nanoseconds
}
```

Call `Circumambient --help` to see all of the options.
