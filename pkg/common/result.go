package common

import (
    "encoding/json"
    "time"

    "github.com/neumannlyu/golog"
)

// 提取的文件信息
type ExFileInfo struct {
    EffectiveNumberOfFiles uint
    KnownNumberOfFiles     uint
    UnknownNumberOfFiles   uint
}

// 时间
type Times struct {
    StartTime time.Time
    EndTime   time.Time
    UsedTime  time.Time
}

type SystemInfo struct {
    OS        string
    CPU       string
    ImageType string
}

// 最终报告结构体
type Report struct {
    Binfile               BinFileInfo
    ExtractedFileInfo     ExFileInfo
    Time                  Times
    SystemInfo            SystemInfo
    KernelVersion         string
    KernelInfo            string
    Kernelvulnerablities  []Vulnerablity
    Programvulnerablities []Vulnerablity
}

func (r Report) ToJson() string {
    data, err := json.Marshal(r)
    if golog.CatchError(err) {
        return ""
    }

    return string(data)
}
