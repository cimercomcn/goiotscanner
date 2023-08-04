package tools

import (
    "fmt"
    "os"
    "os/exec"
    "runtime"

    "github.com/cimercomcn/goiotscanner/pkg/config"
    "github.com/neumannlyu/golog"
)

// 利用binwalk工具提取传入的bin文件。
// binwalk的命令原型为：binwalk -Me example.bin -C out
// @param 待提取的bin文件
// @param 提取后的输出目录
// @return bool
func BinwalkMe(bin_path, out_dir string) bool {
    cfgPtr := *config.GetConfigInstance()
    var cmd *exec.Cmd = nil
    // 执行binwalk
    // 指定要执行的命令和参数
    switch osp := runtime.GOOS; osp {
    case "darwin":
        cmd = exec.Command(
            "sudo",
            "binwalk",
            "-Me",
            "-C", out_dir,
            bin_path,
            "--run-as=root")
        cmd.Stdin = os.Stdin
    case "linux":
        fmt.Printf("out_dir: %v\n", out_dir)
        cmd = exec.Command(
            "binwalk",
            "-Me",
            "-C", out_dir,
            bin_path,
            "--run-as=root")
    }
    cfgPtr.Logs.CommonLog.Debug("Command: " + cmd.String())
    // 执行命令并等待结果
    output, err := cmd.CombinedOutput()
    if golog.CatchError(err) {
        cfgPtr.Logs.CommonLog.Fatal(
            fmt.Sprintln("error executing command: ", err.Error()))
        return false
    }

    // 将结果作为字符串输出
    cfgPtr.Logs.CommonLog.Debug(
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
