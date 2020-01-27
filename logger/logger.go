// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
)

// Logger ...
type Logger struct {
	*logrus.Logger
	Config *config.Config
}

// Tracer ...
func (l *Logger) Tracer(v interface{}) {
	if l.Config.GetBool("TRACER") {
		fmt.Print(v)
	}
}

// New logrus logger
func New(c *config.Config) *Logger {

	logger := &Logger{logrus.New(), c}

	logger.Out = os.Stdout

	if c.GetBool("TRACER") {
		logger.Warn("*** tracer on ***")
	}

	level := strings.ToUpper(c.GetString("LOG_LEVEL"))

	switch level {

	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)

	default:
		logger.SetLevel(logrus.InfoLevel)

	}

	logger.Infof("log level: %s", logger.GetLevel())

	return logger

}
