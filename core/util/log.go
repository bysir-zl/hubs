package util

import "github.com/bysir-zl/bygo/log"

var Logger = log.NewLogger()

func init() {
	Logger.SetTag("hubs")
}