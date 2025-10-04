package models

import (
	"fmt"

	"github.com/rs/zerolog"
)

type zerologBackendLogger struct {
	logger zerolog.Logger
}

func (z *zerologBackendLogger) Debug(v ...any) {
	z.logger.Debug().Msg(fmt.Sprint(v...))
}

func (z *zerologBackendLogger) Debugf(s string, v ...any) {
	z.logger.Debug().Msg(fmt.Sprintf(s, v...))
}

func (z *zerologBackendLogger) Error(v ...any) {
	z.logger.Error().Msg(fmt.Sprint(v...))
}

func (z *zerologBackendLogger) Errorf(s string, v ...any) {
	z.logger.Error().Msg(fmt.Sprintf(s, v...))
}

func (z *zerologBackendLogger) Info(v ...any) {
	z.logger.Info().Msg(fmt.Sprint(v...))
}

func (z *zerologBackendLogger) Infof(s string, v ...any) {
	z.logger.Info().Msg(fmt.Sprintf(s, v...))

}

func (z *zerologBackendLogger) Warn(v ...any) {
	z.logger.Warn().Msg(fmt.Sprint(v...))
}

func (z *zerologBackendLogger) Warnf(s string, v ...any) {
	z.logger.Warn().Msg(fmt.Sprintf(s, v...))
}
