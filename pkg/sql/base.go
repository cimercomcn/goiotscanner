package sql

import "github.com/cimercomcn/goiotscanner/pkg/config"

var _cfgPtr *config.CFG

// sql操作对象
var Isql ISQL

// 导入模块时运行
func init() {
    _cfgPtr = config.GetConfigInstance()
}
