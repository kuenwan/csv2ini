package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"mahonia"
	"os"
	"path/filepath"
	"strings"
)

const (
	FileDir = "./ini/"
)

func checkPathIsExist(path string) bool {
	var exist = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func processFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		fmt.Println(err)
		return err
	}
	if f.IsDir() {
		return nil
	}
	ok := strings.HasSuffix(f.Name(), ".csv")
	if !ok {
		return nil
	}
	println(path)
	cvtRet := convertFile(path)
	if cvtRet == false {
		err1 := fmt.Errorf("convertFile:%v failed", path)
		fmt.Println(err1)
		return err1
	}

	return nil
}

func getFilelist(path string) {
	err := filepath.Walk(path, processFile)
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}

func convertFile(srcFile string) bool {
	fmt.Println("start convert:", srcFile)

	decoder := mahonia.NewDecoder("gb18030")
	if decoder == nil {
		fmt.Println("mahonia newdecoder failed")
		return false
	}

	file, err := os.Open(srcFile)
	if err != nil {
		fmt.Println("open file Error:", err)
		return false
	}
	// 这个方法体执行完成后，关闭文件
	defer file.Close()

	colDesc := []string{}
	colType := []string{}
	colName := []string{}
	fileContent := ""

	row := 0
	reader := csv.NewReader(decoder.NewReader(file))
	for {
		// Read返回的是一个数组，它已经帮我们分割了，
		record, err := reader.Read()
		// 如果读到文件的结尾，EOF的优先级居然比nil还高！
		if err == io.EOF {
			break
		} else if record[0] == "" {
			// 有空行说明配置结束了
			break
		} else if err != nil {
			fmt.Println(fmt.Sprintf("row:%v, error:%v", row, err))
			return false
		} else {
			if row == 0 {
				// 第一行第一列标识该表是否服务器会用到
				tableType := record[0]
				tableType = strings.Trim(tableType, " ")
				tableType = strings.ToLower(tableType)
				if tableType != "common" && tableType != "server" {
					fmt.Println(fmt.Sprintf("tableType:%v", tableType))
					return true
				} else {
					// 继续下一行
					row++
					continue
				}
			} else if row == 1 {
				row++
				continue
			} else if row == 2 {
				// 字段描述行跳过,继续下一行
				colDesc = record
				row++
				continue
			} else if row == 3 {
				// 字段名称行
				colName = record
				row++
				continue
			} else if row == 4 {
				// 字段类型行
				colType = record

				fileContent += fmt.Sprintf(";[%v]\n", colName[0])
				for i := 1; i < len(record); i++ {
					fileContent += fmt.Sprintf(";(%v)%v=%v\n", colType[i], colName[i], colDesc[i])
				}
				fileContent += "\n"

				row++
				continue
			} else {
				// 数据行
				fileContent += fmt.Sprintf("[%v]\n", record[0])
				for i := 1; i < len(record); i++ {
					fileContent += fmt.Sprintf("%v=%v\n", colName[i], record[i])
				}
				fileContent += "\n"
			}
		}

		//fmt.Println(record)

		row++
	}

	//fmt.Println(fmt.Printf("file:%v, content:%v", srcFile, fileContent))
	if fileContent != "" {
		srcFileNameVec := strings.Split(srcFile, ".")
		fileName := srcFileNameVec[0] + ".ini"
		fileFullPathName := FileDir + fileName

		//创建文件
		f, err1 := os.Create(fileFullPathName)
		if err1 != nil {
			fmt.Println(fmt.Printf("file:%v, create error:%s", srcFile, err1))
			return false
		}
		defer f.Close()

		//写入文件(字符串)
		_, err1 = io.WriteString(f, fileContent)
		if err1 != nil {
			fmt.Println(fmt.Printf("file:%v, write error:%s", srcFile, err1))
			return false
		}

		fmt.Println(fmt.Sprintf("generate file:%v\n", fileFullPathName))

		iniFilePathSetFile := FileDir + "iniFilePathSet.ini"
		if checkPathIsExist(iniFilePathSetFile) == false {
			f2, err2 := os.Create(iniFilePathSetFile)
			if err2 != nil {
				fmt.Println(fmt.Printf("file:%v, create error:%s", iniFilePathSetFile, err2))
				return false
			}
			defer f2.Close()

			pathSetContent := ""
			pathSetContent += fmt.Sprintf("[%v]\n", strings.ToUpper(srcFileNameVec[0]))
			pathSetContent += fmt.Sprintf("path=%v\n", fileName)
			pathSetContent += "\n"

			_, err2 = io.WriteString(f2, pathSetContent)
			if err2 != nil {
				fmt.Println(fmt.Printf("file:%v, write error:%s", iniFilePathSetFile, err2))
				return false
			}
		} else {
			// 以只写的模式，打开文件
			f2, err2 := os.OpenFile(iniFilePathSetFile, os.O_WRONLY, 0644)
			if err2 != nil {
				fmt.Println(fmt.Printf("file:%v, open error:%s", iniFilePathSetFile, err2))
				return false
			}
			defer f2.Close()

			pathSetContent := ""
			pathSetContent += fmt.Sprintf("[%v]\n", strings.ToUpper(srcFileNameVec[0]))
			pathSetContent += fmt.Sprintf("path=%v\n", fileName)
			pathSetContent += "\n"

			// 查找文件末尾的偏移量
			n, _ := f2.Seek(0, os.SEEK_END)
			// 从末尾的偏移量开始写入内容
			_, err2 = f2.WriteAt([]byte(pathSetContent), n)
			if err2 != nil {
				fmt.Println(fmt.Printf("file:%v, write error:%s", iniFilePathSetFile, err2))
				return false
			}
		}
	}

	//fmt.Println(row)

	return true
}

func main() {

	println("---start---")

	if checkPathIsExist(FileDir) {
		//如果文件夹存在,删除文件夹
		err := os.RemoveAll(FileDir)
		if err != nil {
			fmt.Println(fmt.Printf("dir:%v, remove error:%s", FileDir, err))
			return
		}
	}

	os.Mkdir(FileDir, os.ModePerm)

	getFilelist("./")

	println("---stop---")
}
