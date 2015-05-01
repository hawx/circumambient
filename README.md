# Circumambient

Simple proxy to log details of any request made through it to redis.


## Getting started

Install [redis](http://redis.io/topics/quickstart) and start `redis-server` if
it isn't already running.

Then,

``` bash
$ go get hawx.me/code/circumambient
$ circumambient --help
...
$ circumambient --in localhost:3002 --out localhost:3001
...
```

Now any requests to <localhost:3002> will be passed to <localhost:3001>, and any
redis subscribers* listening on the channel "requests" will get the following
message:

``` json
{
  "method": "...",
  "url": "...",
  "headers": {

  },
  "timestamp": 1397731289,
  "duration": 14331
}
```

\* Subscribers not included.
