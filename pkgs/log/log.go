package log

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gogf/gf/util/gconv"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.lanhuapp.com/gopkgs/config"
)

var logger *logrus.Logger

type Params map[string]interface{}

func init() {
	logger = logrus.StandardLogger()
	setLoggerLevel()
	switch env := os.Getenv("ENV"); env {
	case "local":
		logger.SetFormatter(&logrus.TextFormatter{})
	case "development":
		fallthrough
	case "dev":
		fallthrough
	case "pre-production":
		fallthrough
	case "production":
		fallthrough
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: true,
		})
	}
	logger.Infoln("Logger initialization successful")
}

func setLoggerLevel() {
	if logger != nil {
		config.SetDefault("log_level", "info")
		logLevel := config.GetString("log_level")

		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			panic("init logger module failed: " + err.Error())
		}
		logger.SetLevel(level)
	}
}

func withCommonFields(t string, f ...Params) *logrus.Entry {
	var fields Params
	if len(f) > 0 {
		fields = f[0]
	}
	if len(f) == 0 {
		fields = make(Params)
	}
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		logger.WithField("mod", "log_module").Errorln("get caller failed.")
	}
	caller := runtime.FuncForPC(pc)

	// common field
	fields["mod"] = t
	if _, ok := fields["debug"]; !ok {
		fields["debug"] = fmt.Sprintf("log from [%s#%d], function is [%s]", file, line, caller.Name())
	}

	return logger.WithFields(logrus.Fields(fields))
}

func Debugln(tag string, args ...interface{}) {
	withCommonFields(tag).Debugln(args...)
}
func Debugf(tag string, format string, args ...interface{}) {
	withCommonFields(tag).Debugf(format, args...)
}
func DebuglnWithField(tag string, fields Params, args ...interface{}) {
	withCommonFields(tag, fields).Debugln(args...)
}
func DebugfWithField(tag string, fields Params, format string, args ...interface{}) {
	withCommonFields(tag, fields).Debugf(format, args...)
}

func Infoln(tag string, args ...interface{}) {
	withCommonFields(tag).Infoln(args...)
}
func Infof(tag, format string, args ...interface{}) {
	withCommonFields(tag).Infof(format, args...)
}
func InfolnWithField(tag string, fields Params, args ...interface{}) {
	withCommonFields(tag, fields).Infoln(args...)
}
func InfofWithField(tag string, fields Params, format string, args ...interface{}) {
	withCommonFields(tag, fields).Infof(format, args...)
}

func Warnln(tag string, args ...interface{}) {
	withCommonFields(tag).Warnln(args...)
}
func Warnf(tag, format string, args ...interface{}) {
	withCommonFields(tag).Warnf(format, args...)
}
func WarnlnWithField(tag string, fields Params, args ...interface{}) {
	withCommonFields(tag, fields).Warnln(args...)
}
func WarnfWithField(tag string, fields Params, format string, args ...interface{}) {
	withCommonFields(tag, fields).Warnf(format, args...)
}

func Errorln(tag string, args ...interface{}) {
	withCommonFields(tag).Errorln(args...)
}
func Errorf(tag, format string, args ...interface{}) {
	withCommonFields(tag).Errorf(format, args...)
}
func ErrorlnWithField(tag string, fields Params, args ...interface{}) {
	withCommonFields(tag, fields).Errorln(args...)
}
func ErrorfWithField(tag string, fields Params, format string, args ...interface{}) {
	withCommonFields(tag, fields).Errorf(format, args...)
}

func Fatalln(tag string, args ...interface{}) {
	withCommonFields(tag).Fatalln(args...)
	os.Exit(-1)
}
func Fatalf(tag, format string, args ...interface{}) {
	withCommonFields(tag).Fatalf(format, args...)
	os.Exit(-1)
}
func FatallnWithField(tag string, fields Params, args ...interface{}) {
	withCommonFields(tag, fields).Errorln(args...)
	os.Exit(-1)
}
func FatalfWithField(tag string, fields Params, format string, args ...interface{}) {
	withCommonFields(tag, fields).Errorf(format, args...)
	os.Exit(-1)
}

// ParseParams 将struct/map格式转成 params
func ParseParams(i interface{}) Params {
	p := Params{}

	m := gconv.Map(i, "json")
	for k, v := range m {
		if _, ok := v.([]byte); ok {
			continue
		}
		if v == nil {
			continue
		}
		vs, err := jsoniter.MarshalToString(v)
		if err != nil {
			p[k] = v
		}
		p[k] = vs
	}

	return p
}
