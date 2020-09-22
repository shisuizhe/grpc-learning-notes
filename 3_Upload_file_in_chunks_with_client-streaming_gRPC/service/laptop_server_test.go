package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pcbook/pb"
	"pcbook/sample"
	"testing"
)

// 服务端的直接调用测试
func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid_uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	require.Nil(t, err)

	testCases := []struct {
		name   string
		laptop *pb.Laptop
		store  LaptopStore
		code   codes.Code
	}{
		{
			name: "success_with_id",
			laptop: sample.NewLaptop(),
			store: NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "success_no_id",
			laptop: laptopNoID,
			store: NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "failure_invalid_id",
			laptop: laptopInvalidID,
			store: NewInMemoryLaptopStore(),
			code: codes.InvalidArgument,
		},
		{
			name: "failure_duplicate_id",
			laptop: laptopDuplicateID,
			store: storeDuplicateID,
			code: codes.AlreadyExists,
		},
	}
	for i := range testCases {
		tc := testCases[i] // 将当前测试用例保存到本地变量（这对于避免并发问题非常重要），因为这里要创建多个并行子测试
		t.Run(tc.name, func(t *testing.T) { // 以tc.name作为子测试名称
			t.Parallel() // 使其与其他测试并行运行

			req := &pb.CreateLaptopRequest{Laptop: tc.laptop}
			server := NewLaptopServer(tc.store, nil)
			res, err := server.CreateLaptop(context.Background(), req)
			if tc.code == codes.OK { // tc.code正确情况
				require.NoError(t, err) // 没有错误
				require.NotNil(t, res)	// 结果不为nil
				require.NotEmpty(t, res.Id) // 返回的laptop.Id不为空
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id) // 因为传的req为指针，所以tc.laptop.Id应该被赋值了；它们应该相等
				}
			} else { // tc.code错误情况
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err) // 获取Status对象
				require.True(t, ok) // ok为true
				require.Equal(t, tc.code, st.Code()) // 因为err来自status.Errorf(code, fromat)，所以这里st(status).code()即拿到了err中的code
			}
		})
	}
}














