package jsonlog

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/rs/zerolog"
)

//newJSONLog ...
func newJSONLog(filepath, confpath string, arg ...interface{}) *zerolog.Logger {
	newfilepath, filename := path.Split(filepath)
	configPath, configName := path.Split(confpath)

	l := zerolog.New(newOut(newfilepath, filename, configPath, configName, arg...)).With().Timestamp().Logger()
	return &l
}

//logClose ...
func logClose(sg os.Signal) {

}

// newInput ...
func newOut(newfilepath, filename, configPath, configName string, arg ...interface{}) *writer {
	os.MkdirAll(newfilepath, os.ModePerm)
	runF, err := os.OpenFile(path.Join(newfilepath, filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("Error in Open run file: ", err)
		os.Exit(1)
	}

	pnewW, err := newWriter(configPath, configName+"level", arg...)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	pnewW.run = runF

	return pnewW
}
