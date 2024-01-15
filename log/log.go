package log

import (
	"github.com/cocosip/utils/database"
	ulog "github.com/cocosip/utils/log"
	"github.com/go-kratos/kratos/v2/log"
	glog "gorm.io/gorm/logger"
	"io"
	stdlog "log"
	"time"
)

func NewLogHelper(logger log.Logger, opt *LogOption) *log.Helper {
	level := log.ParseLevel(opt.GetLevel())
	helper := log.NewHelper(
		log.NewFilter(logger,
			log.FilterLevel(level),
			log.FilterKey(opt.GetFilterKeys()...),
		))
	return helper
}

func NewFileLoggerWithOption(filename string, opt *LogOption) io.Writer {
	return ulog.NewFileLogger(
		ulog.WithFilename(filename),
		ulog.WithMaxSize(int(opt.GetFileOption().MaxSize)),
		ulog.WithMaxAge(int(opt.GetFileOption().GetMaxAge())),
		ulog.WithMaxBackups(int(opt.GetFileOption().GetMaxBackups())),
		ulog.WithLocalTime(opt.GetFileOption().GetLocalTime()),
		ulog.WithCompress(opt.GetFileOption().GetCompress()),
		ulog.WithStdout(opt.GetFileOption().GetStdout()),
	)
}

func NewLogger(w io.Writer, id, name, version string, traceId, spanId interface{}) log.Logger {
	logger := log.With(
		log.NewStdLogger(w),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", name,
		"service.version", version,
		"trace.id", traceId,
		"span.id", spanId,
	)
	return logger
}

func newDefaultConfig() *glog.Config {
	c := &glog.Config{
		SlowThreshold:             500 * time.Millisecond,
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  glog.Warn,
		ParameterizedQueries:      true,
	}
	return c
}

func NewGormLogger(w io.Writer, logOpt *LogOption, opts ...database.GormLoggerOption) glog.Interface {
	c := newDefaultConfig()
	c.LogLevel = getGormLogLevel(logOpt.GetLevel())
	for _, opt := range opts {
		opt(c)
	}
	return glog.New(stdlog.New(w, "", 0), *c)
}

func getGormLogLevel(s string) glog.LogLevel {
	level := log.ParseLevel(s)
	switch level {
	case log.LevelInfo, log.LevelDebug:
		return glog.Info
	case log.LevelWarn:
		return glog.Warn
	case log.LevelError:
		return glog.Error
	default:
		return glog.Silent
	}
}
