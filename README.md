# gobinscan

## 导入

···
go get github.com/neumannlyu/gobinscan
···

## 使用方法

```
start := time.Now()
    InitConfig(
        0xffffffff,
        "../zy.bin",
        "postgresql",
        "172.16.5.114",
        "ly",
        "123456",
        "lydb",
        golog.LOGLEVEL_ALL)
    // cfg.BinExtractedDir = "/Users/neumann/MyBak/Cimer/2023/电力物联网平台/代码/123"
    // InitConfig("../zy.bin",
    //     "mysql",
    //     "172.16.2.115:30006",
    //     "root",
    //     "avtNroKYeB6xXrR3",
    //     "asset_map",
    //     golog.LOGLEVEL_ALL)
    fmt.Printf("Run().ToJson(): %v\n", Run().ToJson())
    elapsed := time.Since(start)
    fmt.Printf("代码执行时间为：%s\n", elapsed)
```