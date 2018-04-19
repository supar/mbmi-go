package main

import (
	"testing"
)

type TestingLogWrap struct {
	*testing.T

	skipErrors bool
}

func (this *TestingLogWrap) Critical(v ...interface{}) {
	this.compat(true, v...)
}

func (this *TestingLogWrap) Error(v ...interface{}) {
	this.compat(true, v...)
}

func (this *TestingLogWrap) Debug(v ...interface{}) {
	this.compat(false, v...)
}

func (this *TestingLogWrap) Info(v ...interface{}) {
	this.compat(false, v...)
}

func (this *TestingLogWrap) Notice(v ...interface{}) {
	this.compat(false, v...)
}

func (this *TestingLogWrap) Warn(v ...interface{}) {
	this.compat(false, v...)
}

func (l *TestingLogWrap) compat(err bool, v ...interface{}) {
	switch v[0].(type) {
	case string:
		if !err || (err && l.skipErrors) {
			l.T.Logf(v[0].(string), v[1:]...)
		} else {
			l.T.Errorf(v[0].(string), v[1:]...)
		}
	default:
		if !err || (err && l.skipErrors) {
			l.T.Log(v...)
		} else {
			l.T.Error(v...)
		}
	}
}
