package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init(level int8) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = MarshalStack
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	Logger = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.ErrorLevel)
	Logger = Logger.Level(zerolog.Level(level))
}
