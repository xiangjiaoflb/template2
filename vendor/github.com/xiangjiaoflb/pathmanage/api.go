package pathmanage

import (
	"fmt"
	"os"
	"path"
)

const (
	//CONFPATH 配置文件请放在此目录下
	cONFPATH = "confpath"

	//LOGPATH 日志文件请放在此目录下
	lOGPATH = "logpath"

	//DATAPATH 数据文件请放在此目录下
	dATAPATH = "datapath"
)

// func init() {
// 	err := os.MkdirAll(cONFPATH, os.ModePerm)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	err = os.MkdirAll(lOGPATH, os.ModePerm)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	err = os.MkdirAll(dATAPATH, os.ModePerm)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

//GetCONFPATH 获取配置文件应该存放的文件夹
func GetCONFPATH(servername string) (confpath string) {
	confpath = path.Join(cONFPATH, fmt.Sprintf("%s_conf", servername))
	err := os.MkdirAll(path.Dir(confpath), os.ModePerm)
	if err != nil {
		panic(err)
	}
	return
}

//GetLOGPATH 获取日志文件应该存放的文件夹
func GetLOGPATH(servername string) (logpath string) {
	logpath = path.Join(lOGPATH, fmt.Sprintf("%s_log", servername))
	err := os.MkdirAll(path.Dir(logpath), os.ModePerm)
	if err != nil {
		panic(err)
	}
	return
}

//GetDATAPATH 获取数据文件应该存放的文件夹
func GetDATAPATH(servername string) (datapath string) {
	datapath = path.Join(dATAPATH, fmt.Sprintf("%s_data", servername))
	err := os.MkdirAll(path.Dir(datapath), os.ModePerm)
	if err != nil {
		panic(err)
	}
	return
}
