package utils

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gogf/gf/util/gconv"
	"github.com/pkg/errors"
	"github.com/sony/sonyflake"
)

var (
	baseTime time.Time
	sf       *sonyflake.Sonyflake
)

func init() {
	baseTime, _ = time.Parse("2006-01-02 15:04:05", "2020-07-10 00:00:00")

	sf = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: baseTime,
		MachineID: func() (uint16, error) {
			return uint16(rand.Intn(100000000)), nil
		},
	})
}

type BaseLog interface {
	Log(interface{})
}

// TryTimes 能重试多次的函数容器，在 SafeRun 容器中运行
func TryTimes(f func() error, times int) error {
	var err error
	for i := 0; i < times; i++ {
		SafeRun(nil, func() {
			err = f()
		})

		if err == nil {
			return nil
		}
	}
	return err
}

// SafeRun 能处理 panic 的函数容器
func SafeRun(l BaseLog, f func()) (backErr error) {
	defer func() {
		if err := recover(); err != nil {
			if l != nil {
				l.Log(err)
			}

			msg := GetCallerInfo(15)
			backErr = fmt.Errorf("panic recover with error %v\n stack is :\n %s", err, msg)
		}
	}()

	f()

	return backErr
}

// GetCallerInfo 获取调用栈信息
func GetCallerInfo(depth int) string {
	callerInfo := ""

	for i := 1; i < depth+1; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return callerInfo
		}

		callerInfo += fmt.Sprintf("%d : %s#%d \n", i, file, line)
	}

	return callerInfo
}

// GetProperSize 将一个 byte 单位的大小，变成一个合适单位的大小，例如 2048 => 2 KB
func GetProperSize(l int) string {
	if l < 1024 {
		return fmt.Sprintf("%d Byte", l)
	}
	if l < 1024*1024 {
		kb := float64(l) / 1024.0
		return fmt.Sprintf("%.2f KB", kb)
	}
	mb := float64(l) / 1048576.0
	return fmt.Sprintf("%.2f MB", mb)
}

// HashCode 获取正数的 hash code
func HashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}

	return 0
}

// GraceShutdown 优雅关闭
func GraceShutdown(funcList []func() error, sigs ...os.Signal) (os.Signal, error) {
	if funcList == nil {
		funcList = []func() error{
			func() error { return nil },
		}
	}

	shutdownChan := make(chan os.Signal, 1)

	if len(sigs) > 0 {
		signal.Notify(shutdownChan, sigs...)
	} else {
		signal.Notify(shutdownChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	}

	sig := <-shutdownChan

	var errList error
	for i, f := range funcList {
		err := TryTimes(f, 3)
		if err != nil {
			errList = errors.Wrap(errList, fmt.Sprintf("func exec error : index [%d], message: %s", i, err))
		}
	}

	return sig, errList
}

// RunAfter 延时执行
func RunAfter(f func(), t time.Duration, safeRun bool) error {
	<-time.After(t)
	if safeRun {
		err := SafeRun(nil, f)
		return err
	}

	f()
	return nil
}

// SetMaxProcs 设置最大进程数
// 其实，从go1.5开始就默认使用最大核数的，所以不用设置也ok
func SetMaxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func GetID() string {
	var id uint64
	TryTimes(func() error {
		var err error
		id, err = sf.NextID()
		return err
	}, 3)

	return gconv.String(id)
}
