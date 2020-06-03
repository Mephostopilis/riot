package api

import (
	"context"
	fmt "fmt"

	"github.com/go-kratos/kratos/pkg/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "service.riot"
const target = "direct://default/127.0.0.1:9007" // NOTE: example

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (SsClient, error) {
	client := warden.NewClient(cfg, opts...)
	// cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	cc, err := client.Dial(context.Background(), fmt.Sprintf("etcd://default/%s", AppID))
	// cc, err := client.Dial(context.Background(), target)
	if err != nil {
		return nil, err
	}
	return NewRiotClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc --bm api.proto
