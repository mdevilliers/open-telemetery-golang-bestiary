
Start up the environment

```shell
cd infra/
docker-compose up
```

Start up a backend service

```shell

cd apps/svc-one
go build && ./svc-one
```

Start up a client api 

```shell

cd apps/client-api
go build && ./client-api
```

Hit the web page hosted by the client API

http://0.0.0.0:8777/hello

Launch Jaeger 

http://0.0.0.0:16686/ 

Launch Zipkin

http://0.0.0.0:9411/zipkin

Find your trace.

