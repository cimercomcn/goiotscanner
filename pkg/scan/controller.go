package scan

import (
    "fmt"

    "github.com/cimercomcn/goiotscanner/pkg/common"
    "github.com/cimercomcn/goiotscanner/pkg/config"
    "github.com/cimercomcn/goiotscanner/pkg/tools"

    "os"
    "path/filepath"

    "github.com/cimercomcn/goiotscanner/pkg/sql"
    "github.com/neumannlyu/golog"
)

// 扫描从这里开始
func Start() common.Report {
    // 保存固件文件信息
    _report.Binfile.Name = filepath.Base(_cfgPtr.BinFile)
    if abs, err := filepath.Abs(_cfgPtr.BinFile); !golog.CatchError(err) {
        _report.Binfile.Dir = filepath.Dir(abs)
        _report.Binfile.MD5 = _report.Binfile.GetMD5()
    }

    // 1. 提取文件
    if _cfgPtr.RunModule&config.MODULE_EXTRACT == 1 && !extractBinFile() {
        return _report
    }
    // 2. 检查固件加密情况
    if !checkIsEncrypted() {
        return _report
    }
    // 3. 扫描提取的文件
    scanExtractedFiles(_cfgPtr.BinExtractedDir)
    // 4. 分析内核漏洞
    scanKernelVulnerability(_cfgPtr.BinExtractedDir)
    // 5. 分析（中间件）程序
    scanProgramVulnerability()

    // 资源回收
    sql.Isql.Close()
    printKernelVulnerabilityInfo(_report.Kernelvulnerablities)
    return _report
}

// 检查固件是否被加密
// true: 加密; false: 未加密
func checkIsEncrypted() bool {
    // 判断条件1: 是否有 squashfs-root 目录
    pass1 := true
    fs, err := os.ReadDir(_cfgPtr.BinExtractedDir)
    golog.CatchError(err)

    // 遍历目录中的文件。如果遇到某些文件时，将跳过。
    for _, file := range fs {
        if file.IsDir() && file.Name() == "squashfs-root" {
            pass1 = false
            break
        }
    }

    // 判断2: 正常情况下squashfs-root目录下应该会有多个文件，如果只有一个文件应该就是加密了。
    pass2 := false
    squashfs_root := filepath.Join(_cfgPtr.BinExtractedDir, "squashfs-root")
    fs, err = os.ReadDir(squashfs_root)
    if golog.CatchError(err) || len(fs) <= 1 {
        pass2 = true
    }

    if pass1 || pass2 {
        _cfgPtr.Logs.CommonLog.Fatal("发现固件被加密，请解密后再尝试分析。")
        return false
    } else {
        _cfgPtr.Logs.Pass.Logln("\n\n\n\t\t\t\t固件加密未加密\n\n\n")
        return true
    }
}

// 提取 bin 文件
// @return 是否成功提取出文件
func extractBinFile() bool {
    // 1.提取固件
    _cfgPtr.Logs.CommonLog.Info(
        fmt.Sprintf("[开始提取] %s > %s\n",
            _cfgPtr.BinFile, _cfgPtr.BinExtractedDir))

    if !tools.BinwalkMe(_cfgPtr.BinFile, _cfgPtr.BinExtractedDir) {
        _cfgPtr.Logs.CommonLog.Fatal("提取固件文件失败")
        return false
    } else {
        _cfgPtr.Logs.OK.Logln(
            "[提取完成] 提取的文件保存在 " + _cfgPtr.BinExtractedDir + " 目录下")
    }
    _cfgPtr.BinExtractedDir = filepath.Join(_cfgPtr.BinExtractedDir,
        "_"+filepath.Base(_cfgPtr.BinFile)+".extracted")
    return true
}

// 打印漏洞信息
func printKernelVulnerabilityInfo(
    kernelvulnerabilities []common.Vulnerablity) {
    // 如果该内核版本有问题就输出
    for i, kernelVulnerability := range kernelvulnerabilities {
        fmt.Printf("[%d]发现漏洞【%s】\n", i, kernelVulnerability.ID)
        fmt.Printf("    漏洞类型【%s】\n", kernelVulnerability.Type)
        fmt.Printf("    漏洞描述【%s】\n", kernelVulnerability.Description)
        fmt.Printf("    漏洞等级【%d】\n", kernelVulnerability.Severity)
        fmt.Printf("    影响范围【内核 %s 】\n",
            kernelVulnerability.AffectedVersion)
        fmt.Printf("    修复建议【%s】\n\n", kernelVulnerability.FixSuggestion)
    }
}
