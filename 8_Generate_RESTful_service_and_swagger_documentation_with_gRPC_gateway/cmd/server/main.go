package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"pcbook/pb"
	"pcbook/service"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 20 * time.Minute
)

const (
	caCertFile = "cert/ca-cert.pem"
	serverCertFile = "cert/server-cert.pem"
	serverKeyFile = "cert/server-key.pem"
)

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin", "123", "admin")
	if err != nil {
		return err
	}
	return createUser(userStore, "pd", "123", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = " /pcbook.pbfiles.LaptopService/"

	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"}, // 只有admin才可以调用
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, errors.New("failed to add client CA's certificate")
	}

	// load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert}, // 服务端证书
		// ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientAuth:   tls.NoClientCert,
		ClientCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func runGRPCServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %w", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
		log.Printf("start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	}

	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	reflection.Register(grpcServer)


	return grpcServer.Serve(listener)
}

func runRESTServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
	grpcEndpoint string, // 改进
) error {
	// 创建一个新的Http请求多路复用器
	// 确保其来自 github.com/grpc-ecosystem/grpc-gateway/runtime
	mux := runtime.NewServeMux()

	// 为了方便演示，这里使用grpc.WithInsecure()
	dialOptions := []grpc.DialOption{grpc.WithInsecure()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 现在开始写从REST到gRPC的进程内转换
	// err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOptions) // 改进
	if err != nil {
		return err
	}

	// in-process handler
	// err = pb.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	err = pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOptions) // 改进
	if err != nil {
		return err
	}

	if enableTLS {
		log.Printf("start REST server at %s, TLS = %t", listener.Addr().String(), enableTLS)
		return http.ServeTLS(listener, mux, serverCertFile, serverKeyFile)
	}
	return http.Serve(listener, mux)
}

func main() {
	port := flag.Int("port", 0, "the server port")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	serverType := flag.String("type", "grpc", "type of grpc (grpc/rest)")
	endPoint := flag.String("endpoint", "", "gprc endpoint") // 改进
	flag.Parse()
	log.Printf("start server on port = %d, TLS = %t", *port, *enableTLS)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	if *serverType == "grpc" {
		err = runGRPCServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
	}
	err = runRESTServer(authServer, laptopServer, jwtManager, *enableTLS, listener, *endPoint)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
