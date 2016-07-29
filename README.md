# HTTP Server Golang implementation

This is a respository for learning of HTTP Server.

Use Golang.


## Usage

```shell
~/g/g/s/http-server-go-impl ❯❯❯ go run main.go &

~/g/g/s/http-server-go-impl ❯❯❯ curl -v -L http://127.0.0.1:8080
* Rebuilt URL to: http://127.0.0.1:8080/
*   Trying 127.0.0.1...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> GET / HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: Test HTTP server
< Content-Length:129
<
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Test</title>
</head>
<body>Hello World!!</body>
* Connection #0 to host 127.0.0.1 left intact
```


## TODO

- remove net.Listen method
- corresponds various Request header, cache-control, content-encoding, and other.
- http#Handlerにinterfaceをあわせる
  - ResponseWriter, Requestの実装
- POST method対応(body)
- Content-Type対応(json, xml, etc)
