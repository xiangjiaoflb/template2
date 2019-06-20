// +build !windows,!nacl,!plan9

package jsonlog

import (
	"io"
	"log"
	"log/syslog"
	"os"
	"path"

	"github.com/xiangjiaoflb/funnel"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var consumer []*funnel.Consumer

//newJSONLog ...
func newJSONLog(filepath, confpath string, arg ...interface{}) *zerolog.Logger {
	newfilepath, filename := path.Split(filepath)
	configPath, configName := path.Split(confpath)

	l := newLogger(newfilepath, filename, configPath, configName, arg...)
	return l
}

//logClose ...
func logClose(sg os.Signal) {
	for _, v := range consumer {
		v.HandleSignal(sg)
	}
}

//NewLogger ...""""
func newLogger(filepath, filename, configPath, configName string, arg ...interface{}) *zerolog.Logger {
	logger, err := syslog.New(syslog.LOG_ERR, path.Join(filepath, filename))
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
		os.Exit(1)
	}

	vp := viper.New()
	//配置初始化
	vp.SetConfigName(configName)
	vp.AddConfigPath(configPath + "/")

	vp.SetDefault(funnel.LoggingDirectory, filepath)
	vp.SetDefault(funnel.LoggingActiveFileName, filename)
	vp.SetDefault(funnel.RotationMaxLines, 100000)
	vp.SetDefault(funnel.RotationMaxFileSizeBytes, 100*1024*1024)
	vp.SetDefault(funnel.Gzip, true)

	//读取传入的参数
	if len(arg) != 0 {
		kvmap, _ := arg[0].(map[string]interface{})
		for k, v := range kvmap {
			vp.SetDefault(k, v)
		}
	}

	cfg, reloadChan, outputWriter, err := funnel.GetConfig(vp, logger, path.Join(configPath, configName+".toml"))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	lp := funnel.GetLineProcessor(cfg)
	c := &funnel.Consumer{
		Config:        cfg,
		LineProcessor: lp,
		ReloadChan:    reloadChan,
		Logger:        logger,
		Writer:        outputWriter,
	}

	dbgR, dbgW := io.Pipe()
	go func() {
		err = c.Start(dbgR)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()
	consumer = append(consumer, c)

	pnewW, err := newWriter(configPath, configName+"level", arg...)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	pnewW.run = dbgW

	slog := zerolog.New(pnewW).With().Timestamp().Logger()
	return &slog
}
