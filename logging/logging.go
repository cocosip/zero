package logging

import (
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

func NewLogger(w io.Writer, id string, name string, traceId interface{}, version string, spanId interface{}) log.Logger {
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

type GormLoggerOption func(o *glog.Config)

func WithGormSlowThreshold(duration time.Duration) GormLoggerOption {
	return func(o *glog.Config) {
		o.SlowThreshold = duration
	}
}

func WithGormLogLevel(level glog.LogLevel) GormLoggerOption {
	return func(o *glog.Config) {
		o.LogLevel = level
	}
}

func NewGormLogger(w io.Writer, logOpt *LogOption, opts ...GormLoggerOption) glog.Interface {
	level := getGormLogLevel(logOpt.GetLevel())
	c := glog.Config{
		SlowThreshold:             1000 * time.Millisecond,
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  level,
		ParameterizedQueries:      true,
	}
	for _, o := range opts {
		o(&c)
	}
	return glog.New(stdlog.New(w, "", 0), c)
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
