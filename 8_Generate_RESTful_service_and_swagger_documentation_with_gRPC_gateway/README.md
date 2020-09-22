#### 一、准备

1. 安装grpc-gateway和protoc-gen-swagger工具

    ```
    go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
    go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
    ```

2. 将 github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis 下的 google 文件夹复制到我们项目的proto目录下。

3. 改造 auth_service.proto 和 laptop_service.proto

4. 修改Makefile的gen生成pb文件命令

    ```
    gen:
    	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb --grpc-gateway_out=:pb --swagger_out=:swagger
    ```

5. make gen

#### 二、说明

在新生成的 auth_service.pb.gw.go 中，RegisterAuthServiceHandlerServer 函数用于从REST到gRPC的进程内转换，这意味着我们不需要运行单独的gRPC服务器来处理通过网络调用过来的REST请求，不幸的是，目前进程内转换仅支持一元RPC，对于流式RPC，我们必须使用 RegisterAuthServiceHandlerFromEndpoint 函数，这会将传入的RESTful请求转换为gRPC格式，并在指定的endpoint上调用相应的RPC。其他 *.pb.gw.go 也类似。

#### 三、swagger

可以看到swagger文件夹下生成了很多*.swagger.json文件，但我们只关心auth_service.swagger.json和laptop_service.swagger.json文件，你也可以在Makefile的gen中定义更详细的生成规则，以减少生成不必要的文件。

这两个文件对于我们创建API文档非常有用，我们可以通过[swagger.io](https://swagger.io/)轻松建立API文档。

1. 登陆swagger.io
2. Create New --> Import and Document APi
3. 选择对应的*.swagger.json文件

![](md-imgs\auth-service.png)

![](md-imgs\laptop-service.png)

#### 四、重构cmd/server/main.go代码

#### 五、Makefile

```
rest:
	go run cmd/server/main.go -port 8081 -type rest
```

make server代表启动gRPC服务

make rest代表启动REST服务

#### 六、测试

1. make rest
2. 从我们上传到swagger.io的接口中，可以轻松拿到接口信息
3. 使用postmain请求接口

**一元请求：**

![](D:\Typora Images\README\login-service-rest.png)

请记住，只有REST服务正在运行，因为我们正在使用进程内转换，因此，它适用于一元请求。

**流式请求：**

让我们看看如果尝试调用流式请求会发生什么？

我们请求 http://localhost:8081/v1/laptop/search（这是服务端流式RPC接口），然后返回了错误："streaming calls are not yet supported in in-process transport"，即进程内传输尚不支持流式调用，see [issues](https://github.com/grpc/grpc-go/issues/906) here。

因此，让我们再稍微改进一下 cmd/server/main.go runRESTServer 代码和修改Makefile。

> 再次测试

首先，启动8080端口的gRPC服务，然后启动8081端口的REST服务。

1. make server
2. make client （用于创建laptop，方便再次请求http://localhost:8081/v1/laptop/search时，可以拿到数据）
3. make rest

![](md-imgs\laptop-service-rest.png)

可以看到，找到了一些laptop，如果往下拉，你会发现它们是独立的JSON对象，而不是数组，原因就是因为这是流式传输的结果。因此服务端发送JSON数据，作为多个单独的JSON对象的流。

