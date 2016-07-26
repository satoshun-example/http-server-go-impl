# HTTP Server Golang implementation

This is a respository for learning of HTTP Server.

Use Golang.


## Usage

```shell
go run main.go &

curl -v -L http://localhost:8080
* Rebuilt URL to: http://localhost:8080/
*   Trying ::1...
* Connected to localhost (::1) port 8080 (#0)
> GET / HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
2016/07/25 13:53:38 / HTTP/1.1

< HTTP/1.1 200 OK
< Server: original HTTP server
* no chunk, no close, no size. Assume close to signal end
<
* Closing connection 0
Hello World!!%
```


## TODO

- remove net.Listen method
- corresponds various Request header, cache-control, content-encoding, and other.
- http#Handlerにinterfaceをあわせる
  - ResponseWriter, Requestの実装
- POST method対応(body)
- Content-Type対応(json, xml, etc)
