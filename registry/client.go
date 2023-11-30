package registry

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	stdgrpc "google.golang.org/grpc"
	zerologging "zero/logging"
)

var (
	_ ClientCreator = (*ClientCreateFunc)(nil)
)

type ClientFactory struct {
	reg     FactoryInterface
	log     *log.Helper
	_logger log.Logger
}

type ClientCreator interface {
	Create(conn *stdgrpc.ClientConn) (interface{}, error)
}

type ClientCreateFunc func(conn *stdgrpc.ClientConn) (interface{}, error)

func (f ClientCreateFunc) Create(conn *stdgrpc.ClientConn) (interface{}, error) {
	return f(conn)
}

func NewClientFactory(reg FactoryInterface, logger log.Logger, logOpt *zerologging.LogOption) *ClientFactory {
	return &ClientFactory{
		reg:     reg,
		log:     zerologging.NewLogHelper(logger, logOpt),
		_logger: logger,
	}
}

func (f *ClientFactory) CreateNewClient(serviceName string, creator ClientCreator) (interface{}, func(), error) {
	var closer func()
	var opts []grpc.ClientOption
	dis, err := f.reg.GetDiscovery()
	if err != nil {
		return nil, closer, err
	}

	opts = append(
		opts,
		grpc.WithEndpoint(serviceName),
		grpc.WithDiscovery(dis),
		grpc.WithMiddleware(
			recovery.Recovery(),
			validate.Validator(),
			logging.Client(f._logger),
		),
	)

	conn, err := grpc.DialInsecure(context.Background(), opts...)
	if err != nil {
		return nil, closer, err
	}

	cli, err := creator.Create(conn)
	if err != nil {
		return nil, closer, err
	}
	closer = func() {
		if err = conn.Close(); err != nil {
			f.log.Errorf("close grpc conn error -> %s", err.Error())
		}
	}
	return cli, closer, nil
}
