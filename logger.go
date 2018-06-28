package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

// RFC5424 log message levels.
// 0       Emergency: system is unusable
// 1       Alert: action must be taken immediately
// 2       Critical: critical conditions
// 3       Error: error conditions
// 4       Warning: warning conditions
// 5       Notice: normal but significant condition
// 6       Informational: informational messages
// 7       Debug: debug-level messages
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// LogIface represents all level methods
type LogIface interface {
	Critical(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	LogIfaceInfo
}

// LogIfaceInfo represents informational methods
type LogIfaceInfo interface {
	Warn(v ...interface{})
	Notice(v ...interface{})
	Info(v ...interface{})
	Debug(v ...interface{})
}

// Log is extended information of the base log package
type Log struct {
	*log.Logger

	levelLog int
	syslog   *syslog.Writer
}

// NewLogger returns Log
func NewLogger() (logger *Log) {
	var (
		err  error
		flag = log.Ldate | log.Ltime | log.Lmicroseconds
		prog = programName + " "
	)

	logger = &Log{
		Logger: log.New(os.Stdout, prog, flag),
	}

	if ConsoleLogFlag == 0 {
		if logger.syslog, err = syslog.New(syslog.LOG_NOTICE|syslog.LOG_USER, prog); err != nil {
			logger.Critical(err)
		} else {
			logger.SetOutput(logger.syslog)
		}
	}

	return
}

// Установи уровень логирования
func (l *Log) SetLevel(level int) {
	l.levelLog = level
}

// Поверь вхождение запрашиваемого уровня в допустимую
// границу логирования
func (l *Log) level(level int) bool {
	for i := LevelEmergency; i <= LevelDebug; i++ {
		if i == level {
			if i > l.levelLog {
				return false
			}
		}
	}
	return true
}

// Единая точка обработки входиящих сообщений согласно
// их уровню и установленной границы логирования
// К сообщению добавляется префикс описание уровня сообщения
func (l *Log) print(level int, v []interface{}) {
	var (
		ln int

		msg,
		prefix string
	)

	ln = len(v)

	if ln == 0 {
		return
	}

	prefix = getPrefix(level)

	switch v[0].(type) {
	case string:
		prefix += v[0].(string)
		v = v[1:]

		msg = fmt.Sprintf(prefix, v...)

	default:
		v = append(v[:1], v[0:]...)
		v[0] = prefix

		msg = fmt.Sprint(v...)
	}

	if l.syslog == nil {
		l.Print(msg)

		return
	}

	switch level {
	case LevelEmergency:
		l.syslog.Emerg(msg)
	case LevelAlert:
		l.syslog.Alert(msg)
	case LevelCritical:
		l.syslog.Crit(msg)
	case LevelError:
		l.syslog.Err(msg)
	case LevelWarning:
		l.syslog.Warning(msg)
	case LevelNotice:
		l.syslog.Notice(msg)
	case LevelInformational:
		l.syslog.Info(msg)
	case LevelDebug:
		l.syslog.Debug(msg)
	}
}

// Префик описание числового значения уровня
func getPrefix(level int) string {
	var prefix = "error"

	switch level {
	case LevelEmergency:
		prefix = "emergency"
	case LevelAlert:
		prefix = "alert"
	case LevelCritical:
		prefix = "critical"
	case LevelError:
		prefix = "error"
	case LevelWarning:
		prefix = "warning"
	case LevelNotice:
		prefix = "notice"
	case LevelInformational:
		prefix = "info"
	case LevelDebug:
		prefix = "debug"
	}

	return "[" + prefix + "] "
}

func (l *Log) Emergency(v ...interface{}) {
	l.print(LevelEmergency, v)
	os.Exit(1)
}

func (l *Log) Alert(v ...interface{}) {
	l.print(LevelAlert, v)
}

func (l *Log) Critical(v ...interface{}) {
	l.print(LevelCritical, v)
	os.Exit(1)
}

func (l *Log) Error(v ...interface{}) {
	l.print(LevelError, v)
}

func (l *Log) Warn(v ...interface{}) {
	l.print(LevelWarning, v)
}

func (l *Log) Notice(v ...interface{}) {
	l.print(LevelNotice, v)
}

func (l *Log) Info(v ...interface{}) {
	l.print(LevelInformational, v)
}

func (l *Log) Debug(v ...interface{}) {
	l.print(LevelDebug, v)
}
