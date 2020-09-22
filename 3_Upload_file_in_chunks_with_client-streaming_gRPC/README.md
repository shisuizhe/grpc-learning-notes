1.在pcbook3/cmd/client/main.go uploadImage 设置超时

```
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

2.在pcbook3/service/laptop_server.go UploadImage 模拟了超时

```
// 模拟超时：假设服务端以某种方式正在非常缓慢地写入数据
time.Sleep(time.Second)
```

3.make server 启动服务端
4.make client 启动客户端

###### 输出如下：

> 服务端

```
...
2020/09/21 14:58:55 receive a chunk with size: 1024
2020/09/21 14:58:56 waiting to receive more data
2020/09/21 14:58:56 rpc error: code = Unknown desc = cannot receive chunk data: rpc error: code = DeadlineExceeded desc = context deadline exceeded
```

> 客户端：

```
2020/09/21 14:58:56 cannot send chunk to server: EOF
exit status 1
```

可以看到：5秒后，我们在服务端上看到了错误日志，错误代码为Unknown，并且还包含DeadlineExceeded错误。

> ==解决==

在调用流接收之前检查ctx错误

pcbook3/service/laptop_server.go UploadImage

```
// check context error
if err := contextError(stream.Context()); err != nil {
return nil
}
```

**复现**

> 服务端

```
2020/09/21 15:15:39 waiting to receive more data
2020/09/21 15:15:39 receive a chunk with size: 1024
2020/09/21 15:15:40 rpc error: code = Canceled desc = request is canceled
```

可以看到：现在在服务端，我们看到了更好的错误日志。

