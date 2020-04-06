package mqengine

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

var logger *zap.Logger
var logOnce sync.Once

func GetLog() *zap.Logger {
	logOnce.Do(func() {
		logger = NewLog()
	})
	return logger
}
func NewLog() *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   config.Log.Filename,   // 日志文件路径
		MaxSize:    config.Log.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: config.Log.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     config.Log.MaxAge,     // 文件最多保存多少天
		Compress:   config.Log.Compress,   // 是否压缩
	}
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	if config.Log.Level == LevelDebug {
		atomicLevel.SetLevel(zap.DebugLevel)
	} else if config.Log.Level == LevelInfo {
		atomicLevel.SetLevel(zap.InfoLevel)
	} else {
		atomicLevel.SetLevel(zap.WarnLevel)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	// filed := zap.Fields(zap.String("serviceName", "serviceName"))
	// 构造日志
	// logger := zap.New(core, caller, development, filed)
	logger := zap.New(core, caller, development)

	logger.Info("log 初始化成功")
	return logger
}
