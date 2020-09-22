package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

// AuthInterceptor is a client interceptor for authentication
type AuthInterceptor struct {
	authClient *AuthClient
	authMethods map[string]bool
	accessToken string
}

// 启动一个单独的goroutine，定期调用登陆API以获取新的访问令牌
// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(
	authClient *AuthClient,
	authMethods map[string]bool,
	refreshDuration time.Duration, // 用于刷新令牌持续时间（多久调用一次登陆API来获取一个新令牌）
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}
	return interceptor, nil
}

// 添加拦截器以将令牌附加到请求上下文中
// Unary returns a client interceptor to authencicate unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log.Printf("--> unary interceptor: %s", method)

		if interceptor.authMethods[method] {
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}

// Stream returns a client interceptor to authencicate stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %s", method)

		if interceptor.authMethods[method] {
			return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (interceptor *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := interceptor.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration // 使用一个等待变量来存储，刷新令牌之前，我们需要等待多久时间
		for {
			time.Sleep(wait)
			err = interceptor.refreshToken()
			if err != nil { // 如果发生错误，我们应该只等一小段时间，假设是1秒，然后重试
				wait = time.Second
			} else { // 如果没有错误，那么我们绝对应该等待刷新持续时间
				wait = refreshDuration
			}
		}
	}()

	return nil
}

func (interceptor *AuthInterceptor) refreshToken() error {
	accessToken, err := interceptor.authClient.Login()
	if err != nil {
		return err
	}

	interceptor.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}