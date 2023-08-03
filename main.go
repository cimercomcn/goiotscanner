package gobinscan

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/cimercomcn/goiotscanner/pkg/common"
    "github.com/cimercomcn/goiotscanner/pkg/config"
    "github.com/cimercomcn/goiotscanner/pkg/scan"
    "github.com/cimercomcn/goiotscanner/pkg/sql"
    "github.com/cimercomcn/goiotscanner/pkg/tools"
    "github.com/neumannlyu/golog"
)

// 全局配置对象
var _cfgPtr *config.CFG

// first call
// configuration initial
func InitConfig(
    runModule uint,
    binfile string,
    databasePlatform string,
    databaseHost string,
    databaseUser string,
    databasePassword string,
    databaseName string,
    loglevel int) *config.CFG {
    fmt.Println(`
▄████  ▒█████   ██▓ ▒█████  ▄▄▄█████▓  ██████  ▄████▄   ▄▄▄       ███▄    █  ███▄    █ ▓█████  ██▀███  
██▒ ▀█▒▒██▒  ██▒▓██▒▒██▒  ██▒▓  ██▒ ▓▒▒██    ▒ ▒██▀ ▀█  ▒████▄     ██ ▀█   █  ██ ▀█   █ ▓█   ▀ ▓██ ▒ ██▒
▒██░▄▄▄░▒██░  ██▒▒██▒▒██░  ██▒▒ ▓██░ ▒░░ ▓██▄   ▒▓█    ▄ ▒██  ▀█▄  ▓██  ▀█ ██▒▓██  ▀█ ██▒▒███   ▓██ ░▄█ ▒
░▓█  ██▓▒██   ██░░██░▒██   ██░░ ▓██▓ ░   ▒   ██▒▒▓▓▄ ▄██▒░██▄▄▄▄██ ▓██▒  ▐▌██▒▓██▒  ▐▌██▒▒▓█  ▄ ▒██▀▀█▄  
░▒▓███▀▒░ ████▓▒░░██░░ ████▓▒░  ▒██▒ ░ ▒██████▒▒▒ ▓███▀ ░ ▓█   ▓██▒▒██░   ▓██░▒██░   ▓██░░▒████▒░██▓ ▒██▒
░▒   ▒ ░ ▒░▒░▒░ ░▓  ░ ▒░▒░▒░   ▒ ░░   ▒ ▒▓▒ ▒ ░░ ░▒ ▒  ░ ▒▒   ▓▒█░░ ▒░   ▒ ▒ ░ ▒░   ▒ ▒ ░░ ▒░ ░░ ▒▓ ░▒▓░
    ░   ░   ░ ▒ ▒░  ▒ ░  ░ ▒ ▒░     ░    ░ ░▒  ░ ░  ░  ▒     ▒   ▒▒ ░░ ░░   ░ ▒░░ ░░   ░ ▒░ ░ ░  ░  ░▒ ░ ▒░
░ ░   ░ ░ ░ ░ ▒   ▒ ░░ ░ ░ ▒    ░      ░  ░  ░  ░          ░   ▒      ░   ░ ░    ░   ░ ░    ░     ░░   ░ 
        ░     ░ ░   ░      ░ ░                 ░  ░ ░            ░  ░         ░          ░    ░  ░   ░     
                                                ░                                                  ░ v0.0.1 
                                                                                                     
    `)

    // set log level, need run it before GetConfigInstance function.
    golog.SetLogLevel(loglevel)
    // 新建一个配置实例
    _cfgPtr = config.GetConfigInstance()
    _cfgPtr.RunModule = runModule
    // 设置二进制固件文件的路径。展开后路径也基于这个路径设定。
    _cfgPtr.BinFile = binfile
    // 设置数据库信息
    _cfgPtr.DB.Platform = databasePlatform
    _cfgPtr.DB.Host = databaseHost
    _cfgPtr.DB.User = databaseUser
    _cfgPtr.DB.Password = databasePassword
    _cfgPtr.DB.Name = databaseName

    // 检查运行环境
    if !checkEnv() {
        _cfgPtr.Logs.CommonLog.Fatal("检查运行环境失败，运行失败")
        os.Exit(0)
    }
    _cfgPtr.Logs.OK.Logln("检查运行环境完成")

    return _cfgPtr
}

func Run() common.Report {
    return scan.Start()
}

// 初始化。检查参数和runtime
// @return bool
func checkEnv() bool {
    // 初始化。检查参数和run time
    _cfgPtr.Logs.CommonLog.Info("开始检查运行环境")

    // 读取配置文件
    _cfgPtr.Logs.CommonLog.Info("[1]正在加载配置文件")
    if !checkConfig() {
        _cfgPtr.Logs.CommonLog.Error("配置检查失败")
        return false
    }
    _cfgPtr.Logs.OK.Logln("加载配置文件完成")

    // 解析命令行参数
    _cfgPtr.Logs.CommonLog.Info("[2]正在解析命令行参数")
    if !checkCmdLine() {
        _cfgPtr.Logs.CommonLog.Fatal("解析命令行参数失败")
        return false
    }
    _cfgPtr.Logs.OK.Logln("解析命令行参数完成")

    _cfgPtr.Logs.CommonLog.Info("[3]正在检查组件")
    if !checkModule() {
        return false
    }
    _cfgPtr.Logs.OK.Logln("检查组件完成")

    return true
}

// 检查命令行参数的正确性
func checkCmdLine() bool {
    // check 1: 检查是否输入了bin文件路径，并且文件的后缀名为'.bin'
    if _cfgPtr.BinFile == "" || !strings.HasSuffix(_cfgPtr.BinFile, ".bin") {
        flag.PrintDefaults()
        return false
    }

    // check 2: 检查bin文件是否存在
    _, err := os.Stat(_cfgPtr.BinFile)
    if os.IsNotExist(err) {
        _cfgPtr.Logs.CommonLog.Fatal("bin file does not exist.")
        return false
    }

    // current work dir
    pwd, err := os.Getwd()
    // new extract dir
    exdir := "_" + filepath.Base(_cfgPtr.BinFile) + ".extracted"
    if golog.CatchError(err) {
        return false
    }
    // 如果未指定提取目录则使用当前目录为提取目录
    if _cfgPtr.RunModule&config.MODULE_EXTRACT == 1 {
        // 需要提取文件

        // 用户没有预先指定提取目录
        if len(_cfgPtr.BinExtractedDir) == 0 {
            // 移除原有的文件夹
            if _, err :=
                os.Stat(filepath.Join(pwd, exdir)); !os.IsNotExist(err) {
                if golog.CatchError(os.RemoveAll(
                    filepath.Join(_cfgPtr.BinExtractedDir,
                        "_"+filepath.Base(_cfgPtr.BinFile)+".extracted"))) {
                    return false
                }
            }
            // 设置默认的提取目录
            _cfgPtr.BinExtractedDir = pwd
        } else { // 指定提取目录
            // 检查指定提取目录是否存在
            if _, err := os.Stat(_cfgPtr.BinExtractedDir); os.IsNotExist(err) {
                if golog.CatchError(err) {
                    _cfgPtr.Logs.CommonLog.Fatal(
                        fmt.Sprintf("Extracted directory %s does not exist.",
                            _cfgPtr.BinExtractedDir))
                    return false
                }

                // 目录存在
                files, err := os.ReadDir(_cfgPtr.BinExtractedDir)
                if golog.CatchError(err) {
                    _cfgPtr.Logs.CommonLog.Fatal(
                        fmt.Sprintf("Error reading directory %s: %v",
                            _cfgPtr.BinExtractedDir, err))
                    return false
                }
                if len(files) != 0 {
                    _cfgPtr.Logs.CommonLog.Fatal(
                        fmt.Sprintf("Extracted directory %s is not empty.\n",
                            _cfgPtr.BinExtractedDir))
                    return false
                }
            }
        }
    } else {
        // 不需要提取文件,直接使用默认的提取路径
        // 提取路径 = 当前路径 + _{binfilename}.extracted
        _cfgPtr.BinExtractedDir = filepath.Join(pwd, exdir)
    }
    return true
}

// 检查必要组件是否已经安装
func checkModule() bool {
    // 1. 检查binwalk是否已经安装
    if _cfgPtr.RunModule&config.MODULE_EXTRACT == 1 {
        if tools.IsInstalledBinwalk() {
            _cfgPtr.Logs.OK.Logln("binwalk")
        } else {
            _cfgPtr.Logs.CommonLog.Fatal("binwalk is not installed")
            return false
        }
    } else {
        _cfgPtr.Logs.CommonLog.Info("skip check binwalk module.")
    }

    // 2. 检查数据库连接
    switch _cfgPtr.DB.Platform {
    case "postgresql":
        sql.Isql = &sql.PostgresSQL{
            DBPtr: nil,
        }
        // 打开数据库
        if !sql.Isql.Open(
            _cfgPtr.DB.Host,
            _cfgPtr.DB.User,
            _cfgPtr.DB.Password,
            _cfgPtr.DB.Name) {
            _cfgPtr.Logs.CommonLog.Fatal("连接数据库失败")
            return false
        }
    case "mysql":
        sql.Isql = &sql.MySQL{
            DBPtr: nil,
        }
        // 打开数据库
        if !sql.Isql.Open(
            _cfgPtr.DB.Host,
            _cfgPtr.DB.User,
            _cfgPtr.DB.Password,
            _cfgPtr.DB.Name) {
            _cfgPtr.Logs.CommonLog.Fatal("连接数据库失败")
            return false
        }
    default:
        _cfgPtr.Logs.CommonLog.Fatal("No Suppoted Platform!!!")
        return false
    }

    _cfgPtr.Logs.OK.Logln("连接数据库")
    return true
}

// 检查配置文件
func checkConfig() bool {
    if _cfgPtr.DB.Host == "" || _cfgPtr.DB.Name == "" ||
        _cfgPtr.DB.Password == "" || _cfgPtr.DB.User == "" {
        return false
    }
    if len(_cfgPtr.ScanPolicy.SkipCustomDirs) == 0 {
        return false
    }
    if len(_cfgPtr.CompressSuffix) == 0 {
        return false
    }
    return true
}
