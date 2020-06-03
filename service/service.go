package service

import (
	"context"
	"fmt"

	pb "github.com/bilibili/kratos/app/interface/passport/api"
	"github.com/bilibili/kratos/app/interface/passport/internal/dao"
	ssopb "github.com/bilibili/kratos/app/service/sso/api"
	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/bilibili/kratos/pkg/conf/paladin"
)

// Service service.
type Service struct {
	ac        *paladin.Map
	dao       dao.Dao
	ssoClient ssopb.SsoClient
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	var (
		rpcClient warden.ClientConfig
		ct        paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&rpcClient); err != nil {
		return
	}

	ssoClient, err := ssopb.NewClient(&rpcClient)
	if err != nil {
		return
	}
	s = &Service{
		ac:        &paladin.TOML{},
		dao:       d,
		ssoClient: ssoClient,
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	return
}

// SayHello grpc demo func.
func (s *Service) SayHello(ctx context.Context, req *pb.HelloReq) (reply *empty.Empty, err error) {
	reply = new(empty.Empty)
	fmt.Printf("hello %s", req.Name)
	return
}

// SayHelloURL bm demo func.
func (s *Service) SayHelloURL(ctx context.Context, req *pb.HelloReq) (reply *pb.HelloResp, err error) {
	reply = &pb.HelloResp{
		Content: "hello " + req.Name,
	}
	fmt.Printf("hello url %s", req.Name)
	return
}

func (s *Service) HeartBeat(ctx context.Context, req *HeartReq) (resp *Reply, err error) {
	return
}

func (s *Service) DocInx(ctx context.Context, req *DocReq) (resp *Reply, err error) {
	return
}

func (s *Service) Delete(ctx context.Context, req *DeleteReq) (resp *Reply, err error) {
	return
}

func (s *Service) Search(ctx context.Context, req *SearchReq) (resp *SearchReply, err error) {
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
