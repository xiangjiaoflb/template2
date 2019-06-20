package jsonlog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Writer ...
type writer struct {
	run         io.Writer
	minLogLevel zerolog.Level
}

func (pointer *writer) Write(p []byte) (n int, err error) {
	//错误在zerolog中有打印
	return pointer.run.Write(p)
}

func (pointer *writer) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	//日志级别大于或等于最小日志级别才打印
	if level >= pointer.minLogLevel {
		return pointer.Write(p)
	}
	return
}

//newWriter ...
func newWriter(configPath, configName string, arg ...interface{}) (*writer, error) {
	//创建实例
	var newW writer

	//创建配置实例
	vp := viper.New()
	//配置初始化
	vp.SetConfigName(configName)
	vp.AddConfigPath(configPath + "/")

	//设置默认配置
	vp.SetDefault(LogLevel, 0)

	//读取传入的参数
	if len(arg) != 0 {
		kvmap, _ := arg[0].(map[string]interface{})
		for k, v := range kvmap {
			vp.SetDefault(k, v)
		}
	}

	//读配置
	err := vp.ReadInConfig()
	if err != nil {
		if vp.ConfigFileUsed() != "" {
			//文件在使用当中则返回错误
			return nil, err
		}
		//没有配置文件则写配置文件
		//创建文件夹
		err = os.MkdirAll(path.Dir(configPath), os.ModePerm)
		if err != nil {
			return nil, err
		}

		//创建配置文件并写入内容
		err = ioutil.WriteFile(path.Join(configPath, configName+".toml"), []byte(fmt.Sprintf("#DebugLevel 0\n#%s=0\n#InfoLevel 1\n#%s=1\n#WarnLevel 2\n#%s=2\n#ErrorLevel 3\n#%s=3\n#FatalLevel 4\n#%s=4\n#PanicLevel 5\n#%s=5\n#NoLevel 6\n#%s=6\n#Disabled 7\n#%s=7\n%s=%d", LogLevel, LogLevel, LogLevel, LogLevel, LogLevel, LogLevel, LogLevel, LogLevel, LogLevel, vp.GetInt(LogLevel))), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// Validate
	if err := validateConfig(vp); err != nil {
		return nil, err
	}

	newW.minLogLevel = zerolog.Level(vp.GetInt(LogLevel))

	//监听配置文件
	vp.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			// Validate
			if err := validateConfig(vp); err != nil {
				log.Println(err)
				return
			}
			newW.minLogLevel = zerolog.Level(vp.GetInt(LogLevel))
		}
	})

	vp.WatchConfig()

	return &newW, nil
}

//判断配置文件是否正确
func validateConfig(v *viper.Viper) error {
	//如果一个数字的字符串转成int类型时出错，那就配置有问题
	_, err := strconv.Atoi(v.GetString(LogLevel))
	if err != nil {
		return err
	}

	return nil
}
