package proto

import (
	"context"

	"google.golang.org/grpc"
)

type SimpleClient interface {
	GetKey(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (*SimpleResponse, error)
}

type simpleClient struct {
	cc *grpc.ClientConn
}

func NewSimpleClient(cc *grpc.ClientConn) SimpleClient {
	return &simpleClient{cc}
}

func (c *simpleClient) GetKey(ctx context.Context, in *SimpleRequest, opts ...grpc.CallOption) (*SimpleResponse, error) {
	out := new(SimpleResponse)
	err := c.cc.Invoke(ctx, "/proto.Simple/GetKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type SimpleServer interface {
	GetKey(context.Context, *SimpleRequest) (*SimpleResponse, error)
}

func RegisterSimpleServer(s *grpc.Server, srv SimpleServer) {
	s.RegisterService(&_Simple_serviceDesc, srv)
}

func _Simple_Getkey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SimpleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServer).GetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Simple/Route",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServer).GetKey(ctx, req.(*SimpleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Simple_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Simple",
	HandlerType: (*SimpleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetKey",
			Handler:    _Simple_Getkey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sim.proto",
}
