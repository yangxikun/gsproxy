package gsproxy

import (
	"os"

	"github.com/op/go-logging"
)

func InitLog(enableColor bool) {
	var format logging.Formatter
	if enableColor {
		format = logging.MustStringFormatter(
			`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s}%{color:reset} %{message}`,
		)
	} else {
		format = logging.MustStringFormatter(
			`%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s} %{message}`,
		)
	}
	b := logging.NewLogBackend(os.Stderr, "", 0)
	bFormatter := logging.NewBackendFormatter(b, format)
	bLeveled := logging.AddModuleLevel(bFormatter)
	bLeveled.SetLevel(logging.INFO, "")
	logging.SetBackend(bLeveled)
}
