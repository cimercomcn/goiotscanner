package iotscanner

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
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

	if strings.HasSuffix(cfgPtr.BinFile, ".zip") {
		unzipPath := filepath.Dir(cfgPtr.BinFile)
		unzip(cfgPtr.BinFile, unzipPath)
		// 找到bin文件
		binFiles, err := findBinFiles(unzipPath)
		if err != nil {
			fmt.Printf("遍历目录失败: %v\n", err)
			common.GReport.RunningReuslt = common.RR_ERROR_EXTRACTED_BIN
			return false
		}

		// 打印找到的 .bin 文件
		if len(binFiles) == 0 {
			fmt.Println("未找到 .bin 文件")
			return false
		} else {
			fmt.Println("找到以下 .bin 文件:")
			for _, file := range binFiles {
				fmt.Println(file)
				cfgPtr.BinFile = file
			}
		}
	}
	// check 1: 检查是否输入了bin文件路径，并且文件的后缀名为'.bin'
	if cfgPtr.BinFile == "" || !(strings.HasSuffix(cfgPtr.BinFile, ".bin")) {
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

// unzip 解压指定的ZIP文件到目标目录
func unzip(src, destPath string) error {
	fmt.Println("src >>>>>>>> " + src)
	fmt.Println("destPath >>>>>>>> " + destPath)
	// 打开ZIP文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("无法打开ZIP文件: %v", err)
	}
	defer r.Close()

	// 遍历ZIP中的文件
	for _, f := range r.File {
		// 构建目标文件路径
		fPath := filepath.Join(destPath, f.Name)

		// 如果是目录，创建目录
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, os.ModePerm); err != nil {
				return fmt.Errorf("无法创建目录: %v", err)
			}
			continue
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return fmt.Errorf("无法创建文件目录: %v", err)
		}

		// 解压文件
		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("无法创建文件: %v", err)
		}
		defer outFile.Close()

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("无法读取ZIP文件中的内容: %v", err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			return fmt.Errorf("无法写入文件: %v", err)
		}
		rc.Close()
	}
	return nil
}

// findBinFiles 遍历指定目录，找到以 .bin 为后缀的文件
func findBinFiles(dir string) ([]string, error) {
	var binFiles []string

	// 遍历目录及其子目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("无法访问路径 %s: %v", path, err)
		}
		// 检查是否为普通文件且以 .bin 结尾
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".bin") {
			binFiles = append(binFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return binFiles, nil
}
