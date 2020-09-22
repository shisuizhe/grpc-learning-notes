1.在pcbook2/cmd/client/main.go searchLaptop 设置请求超时

```go
// 设置请求超时
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel() // 退出主函数之前调用
```

2.在pcbook2/service/laptop_store.go Search 模拟了超时

```go
time.Sleep(time.Second)
log.Print("check laptop id: ", laptop.GetId())
```

3.make server 启动服务端
4.make client 启动客户端

###### 输出如下：

> 服务端

```
2020/09/21 10:51:06 check laptop id: 73f88a6c-add0-4a9a-abeb-b5844828bc79
2020/09/21 10:51:06 send laptop with id: 73f88a6c-add0-4a9a-abeb-b5844828bc79
2020/09/21 10:51:07 check laptop id: 943cce59-0575-4cf0-a178-82d0c27218f9
2020/09/21 10:51:07 send laptop with id: 943cce59-0575-4cf0-a178-82d0c27218f9
2020/09/21 10:51:08 check laptop id: a64a549d-c34b-4387-9c9e-ccfc877ee81d
2020/09/21 10:51:08 send laptop with id: a64a549d-c34b-4387-9c9e-ccfc877ee81d
2020/09/21 10:51:09 check laptop id: aa2d7444-e1a2-483e-9dfe-f67a15a825bb
2020/09/21 10:51:10 check laptop id: ff22e3ec-a647-4e7b-bf95-464c6abfc1b3
```

> 客户端

```
2020/09/21 10:49:42 cannot receive response: rpc error: code = DeadlineExceeded desc = context deadline exceeded
exit status 1
```

可以看到：在客户端，我们收到了 DeadlineExceeded 代码错误；然而，在服务端，它仍然在check打印。这是多余的，因为客户端已经取消了请求。

>==解决==

pcbook2/service/laptop_store.go Search方法加入context

```go
if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
    log.Print("context is canceled")
    return errors.New("context is canceled")
}
```

pcbook2/service/laptop_server.go  SearchLaptop -> Search

```
err := server.Store.Search(
    stream.Context(), // 从流中获取上下文
    filter,
```

**复现**

> 服务端

```
2020/09/21 11:01:51 check laptop id: d270f4f9-30e3-4d6b-b675-39b46ac6b2c4
2020/09/21 11:01:51 send laptop with id: d270f4f9-30e3-4d6b-b675-39b46ac6b2c4
2020/09/21 11:01:52 check laptop id: c9a48bfd-22b2-4a27-9c7f-cd108b3f280c
2020/09/21 11:01:53 check laptop id: 79b63604-98d8-45d7-8d7f-4f0cdd548b52
2020/09/21 11:01:53 send laptop with id: 79b63604-98d8-45d7-8d7f-4f0cdd548b52
2020/09/21 11:01:54 check laptop id: ed98f91b-d13d-455b-9a2f-b82e57bc7a66
2020/09/21 11:01:55 check laptop id: 9c9c750f-1399-40af-b51b-953fec9166a8
2020/09/21 11:01:55 context is canceled
```

当客户端取消请求时，服务端会立即停止处理其他记录，不会做多余的“动作”。