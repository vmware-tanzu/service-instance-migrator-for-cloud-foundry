/* 
 *  Copyright 2022 VMware, Inc.
 *  
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	Logger *Log
	once   sync.Once
)

func init() {
	Logger = NewLogger()
}

type Log struct {
	*logrus.Logger
	Path string
}

func NewLogger() *Log {
	once.Do(func() {
		Logger = createLogger()
	})

	return Logger
}

func SetLogLevel(debug bool) {
	var level string
	if debug {
		level = logrus.DebugLevel.String()
	} else {
		var ok bool
		level, ok = os.LookupEnv("LOG_LEVEL")
		if !ok {
			level = logrus.InfoLevel.String()
		}
	}

	ll, err := logrus.ParseLevel(level)
	if err != nil {
		ll = logrus.InfoLevel
	}
	NewLogger().SetLevel(ll)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func Printf(format string, args ...interface{}) {
	Logger.Printf(format, args...)
}

func Debugln(args ...interface{}) {
	Logger.Debugln(args...)
}

func Infoln(args ...interface{}) {
	Logger.Infoln(args...)
}

func Warnln(args ...interface{}) {
	Logger.Warnln(args...)
}

func Errorln(args ...interface{}) {
	Logger.Errorln(args...)
}

func Fatalln(args ...interface{}) {
	Logger.Fatalln(args...)
}

func Println(args ...interface{}) {
	Logger.Println(args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

func SetOutput(output io.Writer) {
	log.SetOutput(output)
}

func createLogger() *Log {
	var (
		path string
		ok   bool
	)
	if path, ok = os.LookupEnv("SI_MIGRATOR_LOG_FILE"); !ok {
		path = "/tmp/si-migrator.log"
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{
		CallerPrettyfier: caller(9, false),
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyFile: "caller",
		},
		PrettyPrint: false,
	}
	logger.SetOutput(f)
	logger.SetReportCaller(true)

	return &Log{
		Logger: logger,
		Path:   path,
	}
}

// caller returns string presentation of log caller which is formatted as
// `/path/to/file.go:line_number`. e.g. `/path/to/service-instance-migrator/pkg/cmd/export_space.go:25`
func caller(skip int, removePath bool) func(*runtime.Frame) (string, string) {
	return func(f *runtime.Frame) (function string, file string) {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			if removePath {
				return "", fmt.Sprintf("%s:%d", trimPath(f.File), f.Line)
			}
			return "", fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		if removePath {
			return "", fmt.Sprintf("%s:%d", trimPath(file), line)
		}
		return "", fmt.Sprintf("%s:%d", file, line)
	}
}

func trimPath(file string) string {
	slash := strings.LastIndex(file, "/")
	if slash >= 0 {
		file = file[slash+1:]
	}
	return file
}
