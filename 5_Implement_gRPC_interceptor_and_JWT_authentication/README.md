1. 启动服务器
2. 使用Evans客户端连接它(https://github.com/ktr0731/evans)

```
$ evans -r -p 8080

  ______
 |  ____|
 | |__    __   __   __ _   _ __    ___
 |  __|   ' ' / /  / _. | | '_ '  / __|
 | |____   ' V /  | (_| | | | | | '__ ,
 |______|   '_/    '__,_| |_| |_| |___/

 more expressive universal gRPC client

pcbook.pbfiles@127.0.0.1:8080> show service
+---------------+--------------+---------------------+----------------------+
|    SERVICE    |     RPC      |    REQUEST TYPE     |    RESPONSE TYPE     |
+---------------+--------------+---------------------+----------------------+
| AuthService   | Login        | LoginRequest        | LoginResponse        |
| LaptopService | CreateLaptop | CreateLaptopRequest | CreateLaptopResponse |
| LaptopService | SearchLaptop | SearchLaptopRequest | SearchLaptopResponse |
| LaptopService | UploadImage  | UploadImageRequest  | UploadImageResponse  |
| LaptopService | RateLaptop   | RateLaptopRequest   | RateLaptopResponse   |
+---------------+--------------+---------------------+----------------------+
pcbook.pbfiles@127.0.0.1:8080> service AuthService
pcbook.pbfiles.AuthService@127.0.0.1:8080>
pcbook.pbfiles.AuthService@127.0.0.1:8080> call Login                                 
username (TYPE_STRING) => admin                                                       
password (TYPE_STRING) => 1                                                           
command call: rpc error: code = NotFound desc = incorrect username/password
pcbook.pbfiles.AuthService@127.0.0.1:8080> call Login                                 
username (TYPE_STRING) => admin                                                       
password (TYPE_STRING) => 123                                                         
{                                                                                     
	"accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDA2OTUyMTksInVzZXJu
YW1lIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4ifQ.mUQ8EA70Atn9Vx_U56DSvw6UvZtL64XzZs20KA7mgvE"   
}                                                                                      
```

