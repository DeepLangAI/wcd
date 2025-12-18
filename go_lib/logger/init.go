package logger

import (
	conflib "github.com/DeepLangAI/go_lib/conf"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"github.com/DeepLangAI/go_lib/utillib"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Init(logCfg conflib.Logger) {
	if logCfg.LogLevel == "" || logCfg.LogPath == "" {
		panic("logCfg param empty")
	}
	zapCfg := zap.NewProductionEncoderConfig()
	zapCfg.TimeKey = "time"
	zapCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	zapCfg.MessageKey = "message"
	level, err := zap.ParseAtomicLevel(logCfg.LogLevel)
	if err != nil {
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logPaths := []string{logCfg.LogPath}
	if level.Level() == zap.DebugLevel {
		logPaths = append(logPaths, constslib.LogPath_Stdout)
	}
	logPaths = utillib.DeduplicateString(logPaths)

	cores := make([]hertzzap.CoreConfig, 0)
	for _, path := range logPaths {
		if path == constslib.LogPath_Stdout {
			cores = append(cores, hertzzap.CoreConfig{
				Enc: zapcore.NewConsoleEncoder(zapCfg),
				Ws:  zapcore.AddSync(os.Stdout),
				Lvl: level,
			})
		} else {
			cores = append(cores, hertzzap.CoreConfig{
				Enc: zapcore.NewJSONEncoder(zapCfg),
				Ws:  zapcore.AddSync(getWriter(path)),
				Lvl: level,
			})
		}
	}

	l := hertzzap.NewLogger(
		hertzzap.WithZapOptions(zap.AddCaller()),
		hertzzap.WithZapOptions(zap.AddCallerSkip(3)),
		hertzzap.WithExtraKeys([]hertzzap.ExtraKey{
			constslib.TraceIdKey,
			constslib.OperationIdKey,
		}),
		hertzzap.WithCores(
			cores...,
		),
	)
	hlog.SetLogger(l)
}

// 日志切分
func getWriter(filename string) io.Writer {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		panic(absPath)
	}
	hook, err := rotatelogs.New(
		absPath+"_%Y%m%d.log",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		panic(err)
	}
	return hook
}
