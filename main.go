package iotscanner

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
██████╗  ██████╗  ██╗ ██████╗ ████████╗███████╗ ██████╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██████╗ 
██╔════╝ ██╔═══██╗██║██╔═══██╗╚══██╔══╝██╔════╝██╔════╝██╔══██╗████╗  ██║████╗  ██║██╔════╝██╔══██╗
██║  ███╗██║   ██║██║██║   ██║   ██║   ███████╗██║     ███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
██║   ██║██║   ██║██║██║   ██║   ██║   ╚════██║██║     ██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
╚██████╔╝╚██████╔╝██║╚██████╔╝   ██║   ███████║╚██████╗██║  ██║██║ ╚████║██║ ╚████║███████╗██║  ██║
 ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝    ╚═╝   ╚══════╝ ╚═════╝╚═╝   ╚═╝╚═╝ ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
                                                                                    v0.0.11 beta
                                                                                        `)

    // set log level, need run it before GetConfigInstance function.
    golog.SetLogLevel(loglevel)
    // 新建一个配置实例
    cfgPtr := config.GetConfigInstance()
    cfgPtr.RunModule = runModule
    // 设置二进制固件文件的路径。展开后路径也基于这个路径设定。
    cfgPtr.BinFile = binfile
    // 设置数据库信息
    cfgPtr.DB.Platform = databasePlatform
    cfgPtr.DB.Host = databaseHost
    cfgPtr.DB.User = databaseUser
    cfgPtr.DB.Password = databasePassword
    cfgPtr.DB.Name = databaseName
    return cfgPtr
}

func Run() common.Report {
    cfgPtr := config.GetConfigInstance()
    // 检查运行环境
    if !checkEnv() {
        cfgPtr.Logs.CommonLog.Fatal("检查运行环境失败，运行失败")
        common.GReport.RunningReuslt = common.RR_ERROR_CHECK_ENV
        return common.GReport
    }
    cfgPtr.Logs.OK.Logln("检查运行环境完成")
    return scan.Start()
}

// 初始化。检查参数和runtime
// @return bool
func checkEnv() bool {
    cfgPtr := config.GetConfigInstance()
    // 初始化。检查参数和run time
    cfgPtr.Logs.CommonLog.Info("开始检查运行环境")

    // 读取配置文件
    cfgPtr.Logs.CommonLog.Info("[1]正在加载配置文件")
    if !checkConfig() {
        cfgPtr.Logs.CommonLog.Error("配置检查失败")
        return false
    }
    cfgPtr.Logs.OK.Logln("加载配置文件完成")

    // 解析命令行参数
    cfgPtr.Logs.CommonLog.Info("[2]正在解析命令行参数")
    if !checkCmdLine() {
        cfgPtr.Logs.CommonLog.Fatal("解析命令行参数失败")
        return false
    }
    cfgPtr.Logs.OK.Logln("解析命令行参数完成")

    cfgPtr.Logs.CommonLog.Info("[3]正在检查组件")
    if !checkModule() {
        return false
    }
    cfgPtr.Logs.OK.Logln("检查组件完成")

    return true
}

// 检查配置文件
func checkConfig() bool {
    cfgPtr := config.GetConfigInstance()
    if cfgPtr.DB.Host == "" || cfgPtr.DB.Name == "" ||
        cfgPtr.DB.Password == "" || cfgPtr.DB.User == "" {
        cfgPtr.Logs.CommonLog.Debug("The user name or password is empty.")
        return false
    }
    if len(cfgPtr.ScanPolicy.SkipCustomDirs) == 0 {
        cfgPtr.Logs.CommonLog.Debug("The skip dir is empty.")
        return false
    }
    if len(cfgPtr.CompressSuffix) == 0 {
        cfgPtr.Logs.CommonLog.Debug("The compress suffix is empty.")
        return false
    }
    return true
}

// 检查命令行参数的正确性
func checkCmdLine() bool {
    cfgPtr := config.GetConfigInstance()
    cfgPtr.Logs.CommonLog.Debug("check command line args...")
    // check 1: 检查是否输入了bin文件路径，并且文件的后缀名为'.bin'
    if cfgPtr.BinFile == "" || !strings.HasSuffix(cfgPtr.BinFile, ".bin") {
        cfgPtr.Logs.CommonLog.Fatal("Not a valid bin file.")
        flag.PrintDefaults()
        return false
    }

    // check 2: 检查bin文件是否存在
    _, err := os.Stat(cfgPtr.BinFile)
    if os.IsNotExist(err) {
        cfgPtr.Logs.CommonLog.Fatal("bin file does not exist.")
        return false
    }

    // current work dir
    pwd, err := os.Getwd()
    // new extract dir
    exdir := "_" + filepath.Base(cfgPtr.BinFile) + ".extracted"
    if golog.CatchError(err) {
        return false
    }

    // 需要提取文件
    if cfgPtr.RunModule&config.MODULE_EXTRACT == 1 {
        cfgPtr.Logs.CommonLog.Debug("Need extract bin file.")
        // 用户没有预先指定提取目录
        if len(cfgPtr.BinExtractedDir) == 0 {
            cfgPtr.Logs.CommonLog.Debug(
                "User not specify the extraction directory.")
            // 移除原有的文件夹
            if _, err :=
                os.Stat(filepath.Join(pwd, exdir)); !os.IsNotExist(err) {
                if golog.CatchError(os.RemoveAll(
                    filepath.Join(cfgPtr.BinExtractedDir,
                        "_"+filepath.Base(cfgPtr.BinFile)+".extracted"))) {
                    return false
                }
            }
            // 设置默认的提取目录
            cfgPtr.BinExtractedDir = pwd
        } else { // 指定提取目录
            cfgPtr.Logs.CommonLog.Debug(
                "User specified the extraction directory.")
            // 检查指定提取目录是否存在
            _, err := os.Stat(cfgPtr.BinExtractedDir)
            if os.IsNotExist(err) {
                golog.CatchError(err)
                cfgPtr.Logs.CommonLog.Fatal(
                    fmt.Sprintf(
                        "The extraction directory %s does not exist.",
                        cfgPtr.BinExtractedDir))
                return false
            }

            // 目录存在
            files, err := os.ReadDir(cfgPtr.BinExtractedDir)
            if golog.CatchError(err) {
                return false
            }

            cfgPtr.Logs.CommonLog.Debug(
                fmt.Sprintf("len files: %d", len(files)))
            // The extraction directory must empty.
            for _, filename := range files {
                if strings.HasPrefix(filename.Name(), ".") {
                    continue
                }
                cfgPtr.Logs.CommonLog.Fatal(
                    "The extraction directory must empty.")
                return false
            }
        }
    } else {
        cfgPtr.Logs.CommonLog.Debug("Not need extract bin file.")
        // 不需要提取文件,直接使用默认的提取路径
        // 用户没有预先指定提取目录
        if len(cfgPtr.BinExtractedDir) == 0 {
            cfgPtr.Logs.CommonLog.Debug(
                "User not specify the extraction directory.")
            // 提取路径 = 当前路径 + _{binfilename}.extracted
            cfgPtr.BinExtractedDir = filepath.Join(pwd, exdir)
        } else {
            // 指定提取目录
            cfgPtr.Logs.CommonLog.Debug(
                "User specified the extraction directory.")
            // 检查指定提取目录是否存在
            if _, err := os.Stat(cfgPtr.BinExtractedDir); os.IsNotExist(err) {
                if golog.CatchError(err) {
                    cfgPtr.Logs.CommonLog.Fatal(
                        fmt.Sprintf(
                            "The extraction directory %s does not exist.",
                            cfgPtr.BinExtractedDir))
                    return false
                }

                // 目录存在
                cfgPtr.BinExtractedDir = filepath.Join(
                    cfgPtr.BinExtractedDir, exdir)
            }
        }
    }
    return true
}

// 检查必要组件是否已经安装
func checkModule() bool {
    cfgPtr := config.GetConfigInstance()
    // 1. 检查binwalk是否已经安装
    if cfgPtr.RunModule&config.MODULE_EXTRACT == 1 {
        if tools.IsInstalledBinwalk() {
            cfgPtr.Logs.OK.Logln("binwalk")
        } else {
            cfgPtr.Logs.CommonLog.Fatal("binwalk is not installed")
            return false
        }
    } else {
        cfgPtr.Logs.CommonLog.Info("skip check binwalk module.")
    }

    // 2. 检查数据库连接
    switch cfgPtr.DB.Platform {
    case "postgresql":
        sql.Isql = &sql.PostgresSQL{
            DBPtr: nil,
        }
        // 打开数据库
        if !sql.Isql.Open(
            cfgPtr.DB.Host,
            cfgPtr.DB.User,
            cfgPtr.DB.Password,
            cfgPtr.DB.Name) {
            cfgPtr.Logs.CommonLog.Fatal("连接数据库失败")
            return false
        }
    case "mysql":
        sql.Isql = &sql.MySQL{
            DBPtr: nil,
        }
        // 打开数据库
        if !sql.Isql.Open(
            cfgPtr.DB.Host,
            cfgPtr.DB.User,
            cfgPtr.DB.Password,
            cfgPtr.DB.Name) {
            cfgPtr.Logs.CommonLog.Fatal("连接数据库失败")
            return false
        }
    default:
        cfgPtr.Logs.CommonLog.Fatal("No Suppoted Platform!!!")
        return false
    }

    cfgPtr.Logs.OK.Logln("连接数据库")
    return true
}
