package scan

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"

	"github.com/cimercomcn/goiotscanner/pkg/common"
	"github.com/cimercomcn/goiotscanner/pkg/tools"

	"os"
	"path/filepath"

	"github.com/cimercomcn/goiotscanner/pkg/sql"
	"github.com/neumannlyu/golog"
)

// 扫描从这里开始
func Start() common.Report {
	// 保存固件文件信息

	common.GReport.Binfile.Name = filepath.Base(_cfgPtr.BinFile)

	//if filepath.Ext(_cfgPtr.BinFile) == ".zip" {
	//
	//	fmt.Println(">>>>>>>> " + _cfgPtr.BinFile)
	//	unzipPath := filepath.Dir(_cfgPtr.BinFile)
	//	//unzipPath := filepath.Join(absPath)
	//	fmt.Println(">>>>>>>> " + unzipPath)
	//	err := unzip(_cfgPtr.BinFile, unzipPath)
	//	if err != nil {
	//		fmt.Println(err)
	//		common.GReport.RunningReuslt = common.RR_ERROR_EXTRACTED_BIN
	//		return common.GReport
	//	}
	//
	//	// 找到bin文件
	//	binFiles, err := findBinFiles(unzipPath)
	//	if err != nil {
	//		fmt.Printf("遍历目录失败: %v\n", err)
	//		common.GReport.RunningReuslt = common.RR_ERROR_EXTRACTED_BIN
	//		return common.GReport
	//	}
	//
	//	// 打印找到的 .bin 文件
	//	if len(binFiles) == 0 {
	//		fmt.Println("未找到 .bin 文件")
	//	} else {
	//		fmt.Println("找到以下 .bin 文件:")
	//		for _, file := range binFiles {
	//			fmt.Println(file)
	//		}
	//	}
	//	binfilePath := filepath.Join(unzipPath, binFiles[0])
	//	_cfgPtr.BinFile = binfilePath
	//	common.GReport.Binfile.Name = filepath.Base(_cfgPtr.BinFile)
	//	if abs, err := filepath.Abs(_cfgPtr.BinFile); !golog.CatchError(err) {
	//		common.GReport.Binfile.Dir = filepath.Dir(abs)
	//		//common.GReport.Binfile.MD5 = common.GReport.Binfile.GetMD5()
	//	}
	//	// if unzip end
	//} else {
	//	common.GReport.Binfile.Name = filepath.Base(_cfgPtr.BinFile)
	//	if abs, err := filepath.Abs(_cfgPtr.BinFile); !golog.CatchError(err) {
	//		common.GReport.Binfile.Dir = filepath.Dir(abs)
	//		common.GReport.Binfile.MD5 = common.GReport.Binfile.GetMD5()
	//	}
	//}
	//

	common.GReport.Binfile.Name = filepath.Base(_cfgPtr.BinFile)
	if abs, err := filepath.Abs(_cfgPtr.BinFile); !golog.CatchError(err) {
		common.GReport.Binfile.Dir = filepath.Dir(abs)
		common.GReport.Binfile.MD5 = common.GReport.Binfile.GetMD5()
	}
	// 1. 提取文件
	if /*_cfgPtr.RunModule&config.MODULE_EXTRACT == 1 &&*/ !extractBinFile() {
		common.GReport.RunningReuslt = common.RR_ERROR_EXTRACTED_BIN
		return common.GReport
	}

	// 2. 检查固件加密情况
	if !checkIsEncrypted() {
		common.GReport.RunningReuslt = common.RR_ERROR_ENCRYPTED_BIN
		return common.GReport
	}
	// 3. 扫描提取的文件
	scanExtractedFiles(_cfgPtr.BinExtractedDir)
	// 4. 分析内核漏洞
	scanKernelVulnerability(_cfgPtr.BinExtractedDir)
	// 5. 分析（中间件）程序
	scanProgramVulnerability()

	// 资源回收
	sql.Isql.Close()
	printKernelVulnerabilityInfo(common.GReport.Kernelvulnerablities)
	common.GReport.RunningReuslt = common.RR_PASS
	return common.GReport
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

	fmt.Println("_cfgPtr.BinFile >>>>>>>> " + _cfgPtr.BinFile)
	fmt.Println("_cfgPtr.BinExtractedDir >>>>>>>> " + _cfgPtr.BinExtractedDir)
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
