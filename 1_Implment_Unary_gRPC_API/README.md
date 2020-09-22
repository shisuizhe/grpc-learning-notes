### 情况1

1.在pcbook1/cmd/client/main.go设置请求超时

```go
// 设置请求超时
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel() // 退出主函数之前调用
```

2.在pcbook1/service/laptop_server.go模拟了超时

```go
// 模拟超时：假如在这里进行一些繁忙的处理
time.Sleep(6 * time.Second)
```

3.make server 启动服务端
4.make client 启动客户端

###### 输出如下：

> 服务端

```
2020/09/21 00:16:03 receive a create-loatop request with id: 459f0167-1d11-4cec-91eb-24b98ac461c8
2020/09/21 00:16:09 saved latop with id: 459f0167-1d11-4cec-91eb-24b98ac461c8
```

> 客户端

```
2020/09/21 00:16:03 dail server 0.0.0.0:8080
2020/09/21 00:16:08 cannot create laptop: rpc error: code = DeadlineExceeded desc = context deadline exceeded
exit status 1
```

可以看到：在客户端，我们收到了 DeadlineExceeded 代码错误；然而，在服务端，笔记本电脑仍然被创建和保存了；这可能不是我们想要的行为！如果在笔记本电脑被保存之前取消此次请求，我们也希望服务端不保存请求结果，为了达到这样的效果，我们必须检查服务端的上下文错误。

> ==解决==

```go
if ctx.Err() == context.DeadlineExceeded {
	log.Println("deadline is exceeded")
	return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
}
```

**复现**

> 服务端

```
2020/09/21 00:33:19 receive a create-loatop request with id: 27e99425-0483-4eb6-ac3f-258ac1e5ed3f
2020/09/21 00:33:26 deadline is exceeded
```

非常好，服务端不在保存我们请求的结果了！

### 情况2

如果我们通过中断客户端来取消请求会发生什么呢？

1.启动服务端

2.启动客户端，1秒后，按 ctrl+c 停止请求

###### 输出如下

> 服务端

```
2020/09/21 00:37:38 receive a create-loatop request with id: 4c7355a7-d4d7-4dfa-b6b3-a9a2766b6f02
2020/09/21 00:37:45 saved latop with id: 4c7355a7-d4d7-4dfa-b6b3-a9a2766b6f02
```

> 客户端

```
2020/09/21 00:37:38 dail server 0.0.0.0:8080
exit status 2
```

可以看到：在服务端，客户端发过去的请求被保存了，这也不是我们想要的结果，因为客户端已经取消了请求。

> ==解决==

```
if ctx.Err() == context.Canceled {
	log.Println("request is canceled")
	return nil, status.Error(codes.Canceled, "request is canceled")
}
```

**复现**

> 服务端

```
2020/09/21 00:44:40 receive a create-loatop request with id: 121d5dbf-2b27-40f7-ac33-ceb34fd9f223
2020/09/21 00:44:47 request is canceled
```

非常好，达到我们想要的结果了！
