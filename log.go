package gsproxy

import (
	"os"

	"github.com/op/go-logging"
)

func init() {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s}%{color:reset} %{message}`,
	)
	b := logging.NewLogBackend(os.Stderr, "", 0)
	bFormatter := logging.NewBackendFormatter(b, format)
	bLeveled := logging.AddModuleLevel(bFormatter)
	bLeveled.SetLevel(logging.INFO, "")
	logging.SetBackend(bLeveled)
}
