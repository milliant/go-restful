package log

import (
	stdlog "log" //log是系统包  采用别名 stdlog
	"os"
)

//定义了全局对象 Logger，以及全局方法Print ，Printf
//因此在程序中 直接使用Print 以及Printf 会将打印的内容记录到日志
//但是，问题是这个Logger对象，初始化默认是到os.Stderr  也就是全部输出到 标准错误流

// Logger corresponds to a minimal subset of the interface satisfied by stdlib log.Logger
//这里定义了一个接口，包括2个方法。但是这两个方法不是随便定义的。
//从后文可以看出，StdLogger对象实际上是系统log包中的一个对象，因此这两个方法必须是系统包log中Logger 拥有的方法。
//duck-programming
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

var Logger StdLogger

func init() { //在main函数执行之前，自动执行
	// default Logger
	//stdlog.New  --> log.New()其实这是调用系统的方法，产生一个log对象
	SetLogger(stdlog.New(os.Stderr, "[restful] ", stdlog.LstdFlags|stdlog.Lshortfile))
}

func SetLogger(customLogger StdLogger) {
	Logger = customLogger
}

func Print(v ...interface{}) {
	Logger.Print(v...)
}

func Printf(format string, v ...interface{}) {
	Logger.Printf(format, v...)
}
