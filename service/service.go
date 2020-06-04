package service

import (
	"context"
	"fmt"

	"github.com/go-ego/riot"
	pb "github.com/go-ego/riot/api"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/net/rpc/warden"

	"github.com/golang/protobuf/ptypes/empty"
)

// Service service.
type Service struct {
	ac     *paladin.Map
	engine *riot.Engine
}

// New new a service and return.
func New() (s *Service, err error) {
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

func (s *Service) HeartBeat(ctx context.Context, req *pb.HeartReq) (resp *pb.Reply, err error) {
	return
}

func (s *Service) DocInx(ctx context.Context, req *pb.DocReq) (resp *pb.Reply, err error) {
	return
}

func (s *Service) Delete(ctx context.Context, req *pb.DeleteReq) (resp *pb.Reply, err error) {
	// s.engine.RemoveDoc(req.Doc, false)
	return
}

func (s *Service) Search(ctx context.Context, req *pb.SearchReq) (resp *pb.SearchReply, err error) {
	// var docs types.SearchResp

	// docs = s.engine.Search(types.SearchReq{
	// 	Text: sea.Query,
	// 	// NotUseGse: true,
	// 	DocIds: sea.DocIds,
	// 	Logic:  sea.Logic,
	// 	RankOpts: &types.RankOpts{
	// 		OutputOffset: sea.OutputOffset,
	// 		MaxOutputs:   sea.MaxOutputs,
	// 	}})

	// return docs
	return
}

func (s *Service) WgDist(ctx context.Context, req *pb.WgDistReq) (resp *pb.WgDistResp, err error) {
	return
}

func (s *Service) Kill(ctx context.Context, req *pb.KillReq) (resp *pb.KillResp, err error) {
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

// Close close the resource.
func (s *Service) Close() {
}
