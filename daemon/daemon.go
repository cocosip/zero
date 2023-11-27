package daemon

import (
	ud "github.com/cocosip/utils/daemon"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	_ ud.Service = (*KratosService)(nil)
)

type KratosService struct {
	app *kratos.App
	log *log.Helper
}

func NewKratosService(app *kratos.App, logger log.Logger) *KratosService {
	return &KratosService{
		app: app,
		log: log.NewHelper(logger),
	}
}

func (s *KratosService) Name() string {
	return s.app.Name()
}

func (s *KratosService) Run() error {
	return s.app.Run()
}

func (s *KratosService) HandleError(err error) {
	s.log.Errorf("kratos service <%s> error -> %s", s.app.Name(), err.Error())
}
