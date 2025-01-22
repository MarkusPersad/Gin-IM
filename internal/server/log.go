package server

import (
	"errors"
	"fmt"
	"github.com/cilium/lumberjack/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func init() {
	maxSize, _ := strconv.Atoi(os.Getenv("LOG_MAX_SIZE"))
	maxBackups, _ := strconv.Atoi(os.Getenv("LOG_MAX_BACKUPS"))
	maxAge, _ := strconv.Atoi(os.Getenv("LOG_MAX_AGE"))
	isCompressed, _ := strconv.ParseBool(os.Getenv("LOG_COMPRESSED"))
	lv := os.Getenv("LOG_LEVEL")
	currentDate := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("./logs/log-%s.log", currentDate)
	hook := lumberjack.Logger{
		Filename:   logFileName,
		MaxAge:     maxAge,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		Compress:   isCompressed,
	}
	switch strings.ToLower(lv) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}
	zerolog.TimeFieldFormat = os.Getenv("LOG_TIME_FORMAT")
	log.Logger = log.Logger.With().Caller().Stack().Logger()
	if strings.ToLower(lv) == "debug" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: os.Getenv("LOG_TIME_FORMAT"),
		}
		multi := zerolog.MultiLevelWriter(&hook, consoleWriter)
		log.Logger = log.Output(multi)
	} else {
		log.Logger = log.Output(&hook)
	}
}

func GinLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery
		defer func() {
			cost := time.Since(start)
			log.Logger.Info().
				Int("Status", ctx.Writer.Status()).
				Str("Method", ctx.Request.Method).
				Str("Path", path).
				Str("Query", query).
				Str("IP", ctx.ClientIP()).
				Str("User-Agent", ctx.Request.UserAgent()).
				Str("Errors", ctx.Errors.ByType(gin.ErrorTypePrivate).String()).
				Dur("Cost", cost).Send()
		}()
		ctx.Next()
	}
}

func GinRecovery(stack bool) gin.HandlerFunc {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				if brokenPipe {
					log.Logger.Error().
						Str("Path", ctx.Request.URL.Path).
						Any("Error", err).
						Str("Request", string(httpRequest)).
						Send()
					_ = ctx.Error(err.(error))
					ctx.Abort()
					return
				}
				if stack {
					log.Logger.Error().
						Stack().
						Err(errors.New(string(debug.Stack()))).
						Str("error", "[Recovery from panic]").
						Str("request", string(httpRequest)).
						Send()
				} else {
					log.Logger.Error().
						Str("error", "[Recovery from panic]").
						Any("error", err).
						Str("request", string(httpRequest)).
						Send()
				}
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}
