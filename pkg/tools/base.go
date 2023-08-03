package tools

import "github.com/cimercomcn/goiotscanner/pkg/config"

var _cfgPtr config.CFG

func init() {
    _cfgPtr = *config.GetConfigInstance()
}
