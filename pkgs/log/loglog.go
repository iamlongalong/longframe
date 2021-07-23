package log

import (
	"encoding/json"
	"fmt"
	stdLog "log"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/gogf/gf/util/gconv"
	jsoniter "github.com/json-iterator/go"
)

const (
	LEVEL_DEBUG = "debug"
	LEVEL_INFO  = "info"
	LEVEL_WARN  = "warn"
	LEVEL_ERROR = "error"
	LEVEL_PANIC = "panic"
	LEVEL_FATAL = "fatal"
)

var levelMap = map[string]int{
	LEVEL_DEBUG: 1,
	LEVEL_INFO:  2,
	LEVEL_WARN:  3,
	LEVEL_ERROR: 4,
	LEVEL_PANIC: 5,
	LEVEL_FATAL: 6,
}

var logPool = sync.Pool{
	New: func() interface{} {
		return newLogger()
	},
}

// Formatter 内容格式化
type Formatter interface {
	format(logContent) string
}

// JSONFormatter 格式化为json格式
type JSONFormatter struct{}

func (f *JSONFormatter) format(log logContent) string {
	logBytes, _ := jsoniter.Marshal(log)
	return string(logBytes)
}

// TextFormatter 格式化为 text 格式
type TextFormatter struct{}

func (f *TextFormatter) format(log logContent) string {
	var resStr = ""
	switch log["level"] {
	case LEVEL_INFO:
		resStr += "\033[36mINFO\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	case LEVEL_ERROR:
		resStr += "\033[33;41mERROR\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	case LEVEL_FATAL:
		resStr += "\033[33;41mFATAL\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	case LEVEL_PANIC:
		resStr += "\033[33;41mPANIC\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	case LEVEL_DEBUG:
		resStr += "\033[36mDEBUG\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	case LEVEL_WARN:
		resStr += "\033[33mWARN\033[0m \"" + gconv.String(log["msg"]) + "\"   "
	}
	delete(log, "level")
	delete(log, "msg")

	for k, v := range log {
		if v == nil {
			continue
		}
		resStr += fmt.Sprintf("\033[36m%s\033[0m=%s  ", k, gconv.String(v))
	}
	return resStr
}

var formatter Formatter

func init() {
	if os.Getenv("DEBUG_ENV") == "local" {
		formatter = &TextFormatter{}
	} else {
		formatter = &JSONFormatter{}
	}
}

// SetFormatter 设置 formatter
func SetFormatter(f Formatter) {
	formatter = f
}

var stdOutLocker = sync.Mutex{}

// Log 日志
type Log struct {
	publicFields  fields
	privateFields fields
	pubLock       sync.Locker
	single        bool
	message       string
	mod           string
	logQueue      []logContent
	callerDepth   int
	withCaller    bool
}
type logContent map[string]interface{}

func init() {
	stdLog.SetFlags(0) // 设置前缀为空
}

func newLogContent() logContent {
	return make(map[string]interface{}, 5)
}

// Fields 字段
type fields map[string]interface{}

func clearMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}

// Fields 外部使用
type Fields map[string]interface{}

// New 创建一个新的log实例
func New() *Log {
	var logIns *Log
	var ok bool
	if logIns, ok = logPool.Get().(*Log); !ok {
		logIns = newLogger()
	}
	logIns.single = false
	return logIns
}

// N 创建新的log实例
func N() *Log {
	return New()
}

// NS NewSingle别名，创建单例log
func NS() *Log {
	return NewSingle()
}

// NewSingle 创建单例log
func NewSingle() *Log {
	var logIns *Log
	var ok bool
	if logIns, ok = logPool.Get().(*Log); !ok {
		logIns = newLogger()
	}
	logIns.single = true
	return logIns
}

// newLogger 创建一个新的log实例
func newLogger() *Log {
	fields := make(map[string]interface{}, 5)
	return &Log{
		publicFields: fields,
		mod:          "",
		withCaller:   true,
		pubLock:      &sync.Mutex{},
		callerDepth:  3,
		logQueue:     make([]logContent, 0, 3),
	}
}

// AddField 新增field字段，支持 (k,v) 模式，支持 map， 支持 结构体，支持 json string
// TODO 是否需要错误处理？
func (l *Log) AddField(ifield ...interface{}) *Log {
	l.pubLock.Lock()
	defer l.pubLock.Unlock()

	var err error
	if len(ifield) == 1 {
		mapFields := gconv.Map(ifield[0], "json")
		for k, v := range mapFields {
			l.publicFields[k] = v
		}
	} else if len(ifield) == 2 {
		fieldName := gconv.String(ifield[0])

		reflectVal := reflect.ValueOf(ifield[1])
		reflectKind := reflectVal.Kind()

		switch reflectKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
			l.publicFields[fieldName] = ifield[1]

		default:
			var s string
			s, err = jsoniter.MarshalToString(ifield[1])
			if err != nil {
				l.publicFields[fieldName] = ifield[1]
				break
			}
			l.publicFields[fieldName] = s
		}
	}

	return l
}

// SetAction 设置action
func (l *Log) SetAction(action string) *Log {
	l.pubLock.Lock()
	defer l.pubLock.Unlock()

	l.publicFields["action"] = action
	return l
}

// SetMod 设置 模块名 mod
func (l *Log) SetMod(mod string) *Log {
	l.mod = mod
	return l
}

// SetCallerDepth 设置runtime caller深度
func (l *Log) SetCallerDepth(depth int) *Log {
	l.callerDepth = depth
	return l
}

func (l *Log) SetCaller(withCaller bool) *Log {
	l.withCaller = withCaller
	return l
}

func (l *Log) setCommonFields(fields map[string]interface{}) {
	pc, file, line, ok := runtime.Caller(l.callerDepth)
	if !ok {
		l.pubLock.Lock()
		defer l.pubLock.Unlock()

		l.publicFields["_logmod"] = "get caller error"
		return
	}

	if l.withCaller {
		caller := runtime.FuncForPC(pc)
		fields["debug"] = fmt.Sprintf("log from [%s#%d], function is [%s]", file, line, caller.Name())
	}

	if l.mod != "" {
		fields["mod"] = l.mod
	}
}

func (l *Log) checkSend(level string) bool {
	// 单例模式以及fatal级别立马发送
	switch level {
	case LEVEL_FATAL, LEVEL_PANIC:
		return true
	default:
	}

	return l.single == true
}

func (l *Log) setLogProcess(level string, msg string) {
	newLogCon := newLogContent()
	l.setCommonFields(newLogCon)
	if l.privateFields != nil {
		for k, v := range l.privateFields {
			newLogCon[k] = v
			delete(l.privateFields, k)
		}
	}
	newLogCon["level"] = level
	newLogCon["msg"] = msg
	newLogCon["log_time"] = time.Now().UnixNano()

	// 插入log队列
	l.logQueue = append(l.logQueue, newLogCon)
	// 判断是否需要发送
	if l.checkSend(level) {
		l.Send()
	}
	l.clearMod() // 清除非通用mod
}
func (l *Log) clearMod() {
	l.mod = ""
}

func (l *Log) setPrivateField(f interface{}) *Log {
	fieldMap := gconv.Map(f)
	if l.privateFields == nil {
		l.privateFields = make(map[string]interface{}, 0)
	}
	for k, v := range fieldMap {
		l.privateFields[k] = v
	}
	return l
}

// Infof info级别日志
func (l *Log) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.setLogProcess(LEVEL_INFO, msg)
}

// Infoln info级别日志
func (l *Log) Infoln(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_INFO, msg)
}

// InfolnWithField info级别日志
func (l *Log) InfolnWithField(f interface{}, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Infoln(args...)
	l.callerDepth--
}

// InfofWithField info级别日志
func (l *Log) InfofWithField(f interface{}, format string, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Infof(format, args...)
	l.callerDepth--
}

// Debugf debug级别日志
func (l *Log) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.setLogProcess(LEVEL_DEBUG, msg)
}

// DebuglnWithField debug级别日志
func (l *Log) DebuglnWithField(f interface{}, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Debugln(args...)
	l.callerDepth--
}

// DebugfWithField debug级别日志
func (l *Log) DebugfWithField(f interface{}, format string, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Debugf(format, args...)
	l.callerDepth--
}

// Debugln debug级别日志
func (l *Log) Debugln(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_DEBUG, msg)
}

// WarnfWithField warn级别日志
func (l *Log) WarnfWithField(f interface{}, format string, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Warnf(format, args...)
	l.callerDepth--
}

// WarnfWithField warn级别日志
func (l *Log) WarnlnWithField(f interface{}, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Infoln(args...)
	l.callerDepth--
}

// Errorln error 级别日志
func (l *Log) Errorln(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_ERROR, msg)
}

// Errorf error 级别日志
func (l *Log) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.setLogProcess(LEVEL_ERROR, msg)
}

// ErrorlnWithField error级别日志
func (l *Log) ErrorlnWithField(f interface{}, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Errorln(args...)
	l.callerDepth--
}

// ErrorfWithField error级别日志
func (l *Log) ErrorfWithField(f interface{}, format string, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Errorf(format, args...)
	l.callerDepth--
}

// Fatalln fatal 级别日志
func (l *Log) Fatalln(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_FATAL, msg)
}

// Fatalf fatal 级别日志
func (l *Log) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.setLogProcess(LEVEL_FATAL, msg)
}

// FatallnWithField fatal 级别日志
func (l *Log) FatallnWithField(f interface{}, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Fatalln(args...)
	l.callerDepth--
}

// FatalfWithField fatal 级别日志
func (l *Log) FatalfWithField(f interface{}, format string, args ...interface{}) {
	l.callerDepth++
	l.setPrivateField(f).Fatalf(format, args...)
	l.callerDepth--
}

// LogQueueLen 获取log 长度
func (l *Log) LogQueueLen() int {
	return len(l.logQueue)
}

func format(log interface{}) []byte {
	logStr, _ := json.Marshal(log)
	return logStr
}

func (l *Log) bindBaseFields(fields map[string]interface{}) {
	for k, v := range l.publicFields {
		fields[k] = v
	}
}

// Send 主动打印日志
func (l *Log) Send() {
	sendLog(l)
}

// SendLog 打印日志
func sendLog(l *Log) {
	defer func() {
		clearMap(l.publicFields)
		l.logQueue = make([]logContent, 0, 3)
		l.callerDepth = 3
		l.withCaller = true
		logPool.Put(l)
	}()
	stdOutLocker.Lock()
	defer stdOutLocker.Unlock()
	for _, log := range l.logQueue {
		levelStr := log["level"].(string)
		level := levelMap[levelStr]
		msg := log["msg"].(string)
		if checkLevel(levelStr) {
			l.bindBaseFields(log)
			fmt.Fprintln(os.Stdout, formatter.format(log))
			checkSpecialOption(level, msg)
		}
	}
}

func checkSpecialOption(level int, msg string) {
	checkPanic(level, msg)
	checkExit(level)
}

func checkLevel(levelStr string) bool {
	level := levelMap[levelStr]
	return level >= getLevel()
}

func getLevel() int {
	return levelMap[logger.GetLevel().String()]
}

func checkExit(level int) {
	if level == levelMap[LEVEL_FATAL] {
		os.Exit(1)
	}
}

func checkPanic(level int, msg string) {
	if level == levelMap[LEVEL_PANIC] {
		panic(msg)
	}
}

func (l *Log) Print(args ...interface{}) {
	l.Infoln(args...)
}
func (l *Log) Printf(s string, args ...interface{}) {
	l.Infof(s, args...)
}
func (l *Log) Println(args ...interface{}) {
	l.Infoln(args...)
}

func (l *Log) Fatal(arg ...interface{}) {
	l.Fatalln(arg...)
}

func (l *Log) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_PANIC, msg)
}
func (l *Log) Panicf(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	l.setLogProcess(LEVEL_PANIC, msg)
}
func (l *Log) Panicln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.setLogProcess(LEVEL_PANIC, msg)
}

func (l *Log) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.setLogProcess(LEVEL_WARN, msg)
}
func (l *Log) Warnf(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	l.setLogProcess(LEVEL_WARN, msg)
}
func (l *Log) Warnln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.setLogProcess(LEVEL_WARN, msg)
}
