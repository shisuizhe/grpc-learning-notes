#### 一、gRPC服务使用Nginx负载均衡（全部未开启TLS）

> Makefile

```shell
server1:
	go run cmd/server/main.go -port 9001

server2:
	go run cmd/server/main.go -port 9002
```

> nginx.conf

```shell
worker_processes  1;
error_log  logs/error.log;
events {
    worker_connections  10;
}
http {
    access_log  logs/access.log;
    upstream pcbook_services {
	server 0.0.0.0:9001;
	server 0.0.0.0:9002;
    }
    server {
        listen       8080 http2; # 由于gRPC使用Http2，所以加上

        location / {
	    	grpc_pass grpc://pcbook_services;
        }
    }
}
```

> 测试

1. 启动nginx
2. make server1
3. make server2
4. make client

```shell
# server1
2020/09/22 19:36:33 start server on port = 9001, TLS = false
2020/09/22 19:37:00 --> unary interceptor:  /pcbook.pbfiles.AuthService/Login
2020/09/22 19:37:00 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 19:37:00 receive a create-loatop request with id: 95c8e889-5d5c-41bc-8ab6-c54300e0b5dd
2020/09/22 19:37:00 saved latop with id: 95c8e889-5d5c-41bc-8ab6-c54300e0b5dd
```

```shell
# server2
2020/09/22 19:36:40 start server on port = 9002, TLS = false
2020/09/22 19:37:00 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 19:37:00 receive a create-loatop request with id: 0d56d096-b569-4848-9197-2c75d2b8c8a9
2020/09/22 19:37:00 saved latop with id: 0d56d096-b569-4848-9197-2c75d2b8c8a9
2020/09/22 19:37:00 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 19:37:00 receive a create-loatop request with id: 77b627e6-f16e-4443-8d3b-02dbc0c32850
2020/09/22 19:37:00 saved latop with id: 77b627e6-f16e-4443-8d3b-02dbc0c32850
```

```shell
# nginx access.log
127.0.0.1 - - [22/Sep/2020:19:37:00 +0800] "POST /pcbook.pbfiles.AuthService/Login HTTP/2.0" 200 159 "-" "grpc-go/1.27.0"
127.0.0.1 - - [22/Sep/2020:19:37:00 +0800] "POST /pcbook.pbfiles.LaptopService/CreateLaptop HTTP/2.0" 200 43 "-" "grpc-go/1.27.0"
127.0.0.1 - - [22/Sep/2020:19:37:00 +0800] "POST /pcbook.pbfiles.LaptopService/CreateLaptop HTTP/2.0" 200 43 "-" "grpc-go/1.27.0"
127.0.0.1 - - [22/Sep/2020:19:37:00 +0800] "POST /pcbook.pbfiles.LaptopService/CreateLaptop HTTP/2.0" 200 43 "-" "grpc-go/1.27.0"
```

everything is right, great!

#### 二、gRPC服务使用Nginx负载均衡（nginx和client开启TLS）

> Makefile

```shell
client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls
```

在nginx中创建一个文件夹，用于保存 server-cert.pem、server-key.pem和ca-cert.pem。

```shell
mkdir /opt/nginx/cert
```

将上面三个文件拷贝到cert文件夹里。

> nginx.conf

```shell
server {
    listen       8080 ssl http2; # 在监听命令中添加ssl，就启动了TLS

    # Nginx和gRPC客户端之间的相互TLS
    ssl_certificate ../cert/server-cert.pem; # 为nginx提供服务端证书的位置
    ssl_certificate_key ../cert/server-key.pem; # 为nginx提供服务端私钥的位置

    ssl_client_certificate ../cert/ca-cert.pem; # 告诉nginx客户端CA证书的位置
    ssl_verify_client on; # 告诉nginx验证客户端将发送的证书的真实性

    location / {
        grpc_pass grpc://pcbook_services;
    }
}
```

> 测试

1. 启动nginx
2. make server1
3. make server2

**客户端未以TLS模式运行：make client**

请求失败，因为nginx此时已经启用了TLS模式

```
[root@VM-8-10-centos grpc]# make client
go run cmd/client/main.go -address 0.0.0.0:8080
2020/09/22 20:16:32 dail server = 0.0.0.0:8080, TLS = false
2020/09/22 20:16:32 cannot create auth interceptor: rpc error: code = Unavailable desc = connection closed
exit status 1
make: *** [client] Error 1
```

**客户端以TLS模式运行：make client-tls**

```shell
[root@VM-8-10-centos grpc]# make client-tls
go run cmd/client/main.go -address 0.0.0.0:8080 -tls
2020/09/22 20:19:31 dail server = 0.0.0.0:8080, TLS = true
2020/09/22 20:19:32 token refreshed: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDA3NzgzNzIsInVzZXJuYW1lIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4ifQ.t-CGcBiINsCPGTnWJS7IAUYrAF3NAfYsZ7K-W07ZCE8
2020/09/22 20:19:32 --> unary interceptor: /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 20:19:32 create laptop with id: 1114c531-b2aa-4de9-ad4f-42d1067315ee
2020/09/22 20:19:32 --> unary interceptor: /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 20:19:32 create laptop with id: 9203d006-c56b-43ba-b641-09099b3a9689
2020/09/22 20:19:32 --> unary interceptor: /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 20:19:32 create laptop with id: f5c9c201-830b-491d-8fa8-eba855c1e46c
```

哇，成功了！但是请记住，我们的服务端仍没有在TLS模式下运行。

所以只有客户端和nginx之间的连接是安全的，nginx正在通过一个不安全的连接去连接我们的服务端，一旦nginx从客户端收到加密的数据，它将解密数据，然后再将其转发到后端服务器，因此，请确保nginx和后端服务器位于同一受信任的网络中。

如果它们不在同一受信任的网络中，那么别无选择，只能在后端服务器上启用TLS，并配置nginx以使用它。

#### 三、gRPC服务使用Nginx负载均衡（server、nginx和client都开启TLS）

> Makefile

```shell
server1-tls:
	go run cmd/server/main.go -port 9001 -tls

server2-tsl:
    go run cmd/server/main.go -port 9002 -tls
```

> 测试

1. 启动nginx
2. make server1-tls
3. make server2-tls
4. make client-tls

```shell
go run cmd/client/main.go -address 0.0.0.0:8080 -tls
2020/09/22 20:35:38 dail server = 0.0.0.0:8080, TLS = true
2020/09/22 20:35:38 cannot create auth interceptor: rpc error: code = Unavailable desc = Bad Gateway: HTTP status code 502; transport: received the unexpected content-type "text/html"
exit status 1
```

可以看到，尽管客户端和nginx之间的TLS握手成功，但nginx和我们的后端服务器之间的TLS握手失败。

查看nginx错误日志：

```shell
[root@VM-8-10-centos logs]# cat error.log 
2020/09/22 20:35:38 [error] 11540#0: *1 upstream prematurely closed connection while reading response header from upstream, client: 127.0.0.1, server: , request: "POST /pcbook.pbfiles.AuthService/Login HTTP/2.0", upstream: "grpc://0.0.0.0:9001", host: "0.0.0.0:8080"
```

如日志所示，当nginx与上游服务器对话时，发生了故障。

为了让nginx与后端服务器TLS握手成功，现在，让我们设置一下吧。

修改 cmd/server/main.go loadTLSCredential 函数的代码：

```go
config := &tls.Config{
    Certificates: []tls.Certificate{serverCert},
    // ClientAuth:   tls.RequireAndVerifyClientCert, // <-- not this
    // 意味着我们只使用服务端TLS
    ClientAuth:   tls.NoClientCert, // <---- to this
    ClientCAs: certPool,
}
```

> nginx.conf

```shell
location / {
    # grpc_pass grpc://pcbook_services;
    grpc_pass grpcs://pcbook_services;
}
```

> 重新测试

1. 启动nginx
2. make server1-tls
3. make server2-tls
4. make client-tls

这次成功了。

**如果我们真的想要nginx和上游服务端之间的双向TLS，怎么做？**

1. 将 cmd/server/main.go loadTLSCredential 函数修改的代码改回来。
2. 指示nginx与后端服务器进行双向TLS。

```shell
location / {
	grpc_pass grpcs://pcbook_services;

    # Nginx和gRPC服务端之间的相互TLS
    grpc_ssl_certificate ../cert/server-cert.pem;
    grpc_ssl_certificate_key ../cert/server-key.pem;
}
```

> 再次重新测试

可以发现这次成功了，server、nginx和client都开启TLS，nice！

#### 四、服务分离

从刚才的测试中，我们可以发现，Login请求和Create请求被平均分配到两台server；但是有时候，我们可能想要分离身份验证服务和业务逻辑服务；例如，假设我们希望所有的Login请求都发送到server1，其他请求都发往server2。

让我们配置一下nginx.conf吧：

```shell
http {
    access_log  logs/access.log;
    
    upstream auth_services {
        server 0.0.0.0:9001;
    }

    upstream laptop_services {
        server 0.0.0.0:9002;
    }
    server {
        listen       8080 ssl http2;
        
        ssl_certificate ../cert/server-cert.pem;
        ssl_certificate_key ../cert/server-key.pem;
         
        ssl_client_certificate ../cert/ca-cert.pem;
        ssl_verify_client on;
        
        location /pcbook.pbfiles.AuthService {
            grpc_pass grpcs://auth_services; 
            
            grpc_ssl_certificate ../cert/server-cert.pem;
            grpc_ssl_certificate_key ../cert/server-key.pem;
        }   

        location /pcbook.pbfiles.LaptopService {
            grpc_pass grpcs://laptop_services; 
            
            grpc_ssl_certificate ../cert/server-cert.pem;
            grpc_ssl_certificate_key ../cert/server-key.pem;
        }   
    }   
} 
```

> 测试一下吧

**server1**

```
[root@VM-8-10-centos grpc]# make server1-tls
go run cmd/server/main.go -port 9001 -tls
2020/09/22 21:22:56 start server on port = 9001, TLS = true
2020/09/22 21:23:06 --> unary interceptor:  /pcbook.pbfiles.AuthService/Login
2020/09/22 21:23:36 --> unary interceptor:  /pcbook.pbfiles.AuthService/Login
2020/09/22 21:24:07 --> unary interceptor:  /pcbook.pbfiles.AuthService/Login
```

**server2**

```
[root@VM-8-10-centos grpc]# make server2-tls
go run cmd/server/main.go -port 9002 -tls
2020/09/22 21:23:00 start server on port = 9002, TLS = true
2020/09/22 21:23:07 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 21:23:07 receive a create-loatop request with id: b6d36bf3-55f4-4bac-8c8b-8ef844d21c4a
2020/09/22 21:23:07 saved latop with id: b6d36bf3-55f4-4bac-8c8b-8ef844d21c4a
2020/09/22 21:23:07 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 21:23:07 receive a create-loatop request with id: 8b2a1a65-124d-4a00-b189-142e93bd6720
2020/09/22 21:23:07 saved latop with id: 8b2a1a65-124d-4a00-b189-142e93bd6720
2020/09/22 21:23:07 --> unary interceptor:  /pcbook.pbfiles.LaptopService/CreateLaptop
2020/09/22 21:23:07 receive a create-loatop request with id: 7d184174-2ecf-4fad-91cc-ee427e0fb5b4
2020/09/22 21:23:07 saved latop with id: 7d184174-2ecf-4fad-91cc-ee427e0fb5b4
```

哇，nice！

