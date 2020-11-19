# Opentracing Example

First, clone this repo in your `$GOPATH/github.com/yusufsyaifudin`.

After that, inside the directory `$GOPATH/github.com/yusufsyaifudin/go-opentracing-example` run:

```bash
$ docker-compose up -d #this will run the Jaeger service
$ make run #this will run the example server in port 1323
```

Then use the CURL:

```
curl -I -X GET 'http://localhost:1323/dora-the-explorer?is_rainy_day=true' 
```

You'll get the response like this:

```
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Uber-Trace-Id: 1f43f36f2edd33ed:1f43f36f2edd33ed:0:1
Date: Sat, 08 Dec 2018 06:01:54 GMT
Content-Length: 54
```

There, as you can see there is a header `Uber-Trace-Id` and you can get the trace info in Jaeger using following URL: [http://localhost:16686/trace/1f43f36f2edd33ed](http://localhost:16686/trace/1f43f36f2edd33ed)

## Load testing

For single request:

```sh
$ k6 run load-testing.js
```

For load testing request with 10 virtual users in 30 seconds:

```sh
$ k6 run --vus 10 --duration 30s load-testing.js
```