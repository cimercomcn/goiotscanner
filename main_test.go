package iotscanner

import (
    "fmt"
    "testing"
    "time"

    "github.com/neumannlyu/golog"
)

// 在固件中squashfs-root-0/sbin 目录下nfsroot是什么文件？
func TestMain(m *testing.M) {
    start := time.Now()
    // InitConfig(
    //     0x0,
    //     "../zy.bin",
    //     "postgresql",
    //     "172.16.5.114",
    //     "ly",
    //     "123456",
    //     "lydb",
    //     golog.LOGLEVEL_ALL)
    // cfg.BinExtractedDir = "/Users/neumann/MyBak/Cimer/2023/电力物联网平台/代码/123"
    InitConfig(
        0x0,
        "../zy.bin",
        "mysql",
        "172.16.2.117:3306",
        "root",
        "19UiJ1gq0Cnc5rQG5bo6",
        "mapping",
        golog.LOGLEVEL_ALL)

    fmt.Printf("Run().ToJson(): %v\n", Run().ToJson())
    elapsed := time.Since(start)
    fmt.Printf("代码执行时间为：%s\n", elapsed)
}
