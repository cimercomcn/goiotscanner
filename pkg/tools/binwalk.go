package tools

import (
    "fmt"
    "os"
    "os/exec"
    "runtime"
    "strings"

    "github.com/cimercomcn/goiotscanner/pkg/config"
)

// 利用binwalk工具提取传入的bin文件。
// binwalk的命令原型为：binwalk -Me example.bin -C out
// @param 待提取的bin文件
// @param 提取后的输出目录
// @return bool
func BinwalkMe(bin_path, out_dir string) bool {
    fmt.Printf("out_dir: %v\n", out_dir)
    var cmd *exec.Cmd = nil
    // 执行binwalk
    // 指定要执行的命令和参数
    switch osp := runtime.GOOS; osp {
    case "darwin":
        _cfgPtr = *config.GetConfigInstance()
        cmd = exec.Command("sudo", "binwalk", "-Me", bin_path, "--run-as=root")
        cmd.Stdin = os.Stdin
        // fmt.Printf("_cfgPtr.BinExtractedDir: %v\n", _cfgPtr.BinExtractedDir)
        lastSlashIndex := strings.LastIndex(_cfgPtr.BinExtractedDir, "/")

        if lastSlashIndex != -1 {
            // 使用字符串切片截取最后一个斜杠前面的内容
            _cfgPtr.BinExtractedDir = _cfgPtr.BinExtractedDir[:lastSlashIndex]
            // fmt.Println("Path without last segment:", result)
        }
        // fmt.Printf("_cfgPtr.BinExtractedDir: %v\n", _cfgPtr.BinExtractedDir)
    case "linux":
        cmd = exec.Command("binwalk", "-Me", "-C",
            out_dir, bin_path, "--run-as=root")
    }
    // fmt.Printf("cmd.String(): %v\n", cmd.String())
    // 执行命令并等待结果
    output, err := cmd.CombinedOutput()
    if err != nil {
        _cfgPtr.Logs.CommonLog.Fatal(
            fmt.Sprintln("error executing command:", err.Error()))
        return false
    }

    // 将结果作为字符串输出
    _cfgPtr.Logs.CommonLog.Info(
        fmt.Sprintln(string(output)))
    return true
}

// 检查系统中是否安装了binwalk。
// ! binwalk需要添加到环境变量中。
// @return bool
// @return error
func IsInstalledBinwalk() bool {
    programName := "binwalk"

    // 检查程序是否安装
    cmd := exec.Command("which", programName)
    if err := cmd.Run(); err != nil {
        return false
    } else {
        return true
    }
}
