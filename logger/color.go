package logger

import "github.com/fatih/color"

func InfoC(format string, arg ...interface{}) {
	color.Green(format, arg...)
}

func DebugC(format string, arg ...interface{}) {
	color.Magenta(format, arg...)
}

func WarnC(format string, arg ...interface{}) {
	color.Yellow(format, arg...)
}

func ErrorC(format string, arg ...interface{}) {
	color.Red(format, arg...)
}
