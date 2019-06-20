package funnel

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// XXX: Move it to constants.go if needed
const (
	// config keys
	LoggingDirectory         = "logging.directory"
	LoggingActiveFileName    = "logging.active_file_name"
	LoggingWriteOutput       = "logging.write_output"
	RotationMaxLines         = "rotation.max_lines"
	RotationMaxFileSizeBytes = "rotation.max_file_size_bytes"
	FlushingTimeIntervalSecs = "flushing.time_interval_secs"
	FileRenamePolicy         = "rollup.file_rename_policy"
	MaxAge                   = "rollup.max_age"
	MaxCount                 = "rollup.max_count"
	Gzip                     = "rollup.gzip"
	PrependValue             = "misc.prepend_value"
	Target                   = "target.name"

	MACRO = "{macroConf}"
)

var (
	defaultConf = fmt.Sprintf(`# Sample config file
# The values mentioned are the default values

[logging]
# The directory to store the log files
directory = "%s"
# The name of the current log file
active_file_name = "%s"
# Write standard output
write_output = %s

# File will be rotated whenever any one of these conditions are met
[rotation]
# Max no. of lines beyond which the file will rotate
max_lines = %s # hundred thousand
# Max no. of bytes written to a file beyond which it will rotate
max_file_size_bytes = %s # 5MB

# The time interval after which the buffer will be flushed to the output target.
# For some targets, flushing doesn't make sense. It becomes a no-op then.
# Other targets have in-built flush frequency. It can be configured in that section.
[flushing]
time_interval_secs = %s

[rollup]
# Specify file rename policy.
# Values accepted are
# timestamp - rotated files will be named with the timestamp at the moment of rotation
# serial - rotated files will be named serially in an increasing sequence
file_rename_policy = "%s"
# The maximum age of a file beyond which it will be removed
# Suffix must be either d(days) or h(hours)
max_age = "%s"
# The maximum no. of files to keep in the log directory
# Older files will be deleted first
max_count = %s
# Whether to gzip the rolled over files or not
gzip = %s

[misc]
# Populate the following variable if you want to
# prepend your log line with a predefined text.
# There are some template values you can use too.
# {{.RFC822Timestamp}} expands to a timestamp in RFC822 format
# {{.ISO8601Timestamp}} expands to a timestamp in ISO8601 format
# {{.UnixTimestamp}} expands to a unix epoch timestamp to nanosecond precision
#
# Example -
# prepend_value = "[app_name]- "
# prepend_value = "[app_name] {{.RFC822Timestamp}}- "
prepend_value = "%s"

# Specifies the output target to send the logs to. Uncomment the output you want.
# You can omit this section if you are just logging to files.

# Kafka output example
# [target]
# name = "kafka"
# brokers = ["host1:port", "host2:port"]
# topic = "testtopic"
# clientID = "funnel"
# You need not set the below settings. They will be set to kafka default values if not specified
# flush_frequency_secs = 5 # Best-effort frequency of flushing messages
# batch_size = 10 # Best-effort num of messages to trigger a flush
# max_retries = 3 # The total number of times to retry sending a message
# write_timeout_secs = 30 # How long to wait for a transmit.

# Redis output example
# [target]
# name = "redis"
# host = "localhost:6379"
# password = "" # If no password is set, keep it blank
# channel = "test" # Specify the channel to publish to

# ElasticSearch output example
# P.S. Since ES takes json objects, log lines have to be in json.
# eg lines-
# {"User": "bacon", "Message": "i will be back !"}
# {"User": "tea", "Message": "Open the bifrost !"}
# [target]
# name = "elasticsearch"
# nodes = ["http://host1:port", "http://host2:port"]
# index = "testindex"
# type = "testtype"
# You can set the username and password to blank if you are not using basic auth
# username = "testuser"
# password = "testpass"

# InfluxDB output example
# P.S. InfluxDB has the concept of tags and fields. Log lines have to be in this format -
# {"tags": {"tag1": "value1", "tag2": "other_value1"}, "fields": {"field1": 10, "field2": 20}}
# {"tags": {"tag1": "value2", "tag2": "other_value2"}, "fields": {"field1": 11, "field2": 21}}
# [target]
# name = "influxdb"
# host = "http://localhost:8086" # or "localhost:8089" in case of udp
# db = "testdb" # only valid for http. For udp, database if taken from influxDB config
# protocol = "http" # options are http, udp
# metric = "testmetric"
# username = "testuser"
# password = "testpass"
# time_precision = "s" # options are "ns", "us" (or "µs"), "ms", "s", "m", "h"

# AWS S3 output example
# P.S. Files in s3 are named with the current timestamp
# [target]
# name = "s3"
# bucket = "bucket-name"
# region = "us-west-2"

# NATS output example
# [target]
# You can omit the user and password field if you don't have authentication set up
# name = "nats"
# host = "localhost"
# port = "4222"
# subject = "testsub"
# user = "testuser"
# password = "testpass"
`, MACRO+LoggingDirectory,
		MACRO+LoggingActiveFileName,
		MACRO+LoggingWriteOutput,
		MACRO+RotationMaxLines,
		MACRO+RotationMaxFileSizeBytes,
		MACRO+FlushingTimeIntervalSecs,
		MACRO+FileRenamePolicy,
		MACRO+MaxAge,
		MACRO+MaxCount,
		MACRO+Gzip,
		MACRO+PrependValue)
)

var (
	// ErrInvalidFileRenamePolicy is raised for invalid values to file rename policy
	ErrInvalidFileRenamePolicy = errors.New(FileRenamePolicy + " can only be timestamp or serial")
	// ErrInvalidMaxAge is raised for invalid value in max age - life bad suffixes or no integer value at all
	ErrInvalidMaxAge = errors.New(MaxAge + " must end with either d or h and start with a number")
)

// ConfigValueError holds the error value if a config key contains
// an invalid value
type ConfigValueError struct {
	Key string
}

func (e *ConfigValueError) Error() string {
	return "Invalid config value entered for - " + e.Key
}

// Config holds all the config settings
type Config struct {
	DirName        string
	ActiveFileName string
	WriteOutput    bool

	RotationMaxLines int
	RotationMaxBytes uint64

	FlushingTimeIntervalSecs int

	PrependValue string

	FileRenamePolicy string
	MaxAge           int64
	MaxCount         int
	Gzip             bool

	Target string
}

// GetConfig returns the config struct which is then passed
// to the consumer
func GetConfig(v *viper.Viper, logger *syslog.Writer, confpath string) (*Config, chan *Config, OutputWriter, error) {
	// Set default values. They are overridden by config file values, if provided
	setDefaults(v)
	// Create a chan to signal any config reload events
	reloadChan := make(chan *Config)

	// Find and read the config file
	err := v.ReadInConfig()
	if err != nil {
		if v.ConfigFileUsed() != "" {
			// Return the error only if config file is present
			return nil, reloadChan, nil, err
		} else {
			//没有配置文件则写配置文件
			//创建文件夹
			err = os.MkdirAll(path.Dir(confpath), os.ModePerm)
			if err != nil {
				return nil, reloadChan, nil, err
			}

			if v.GetString(Gzip) != "true" {
				v.SetDefault(Gzip, false)
			}

			if v.GetString(Target) == "" {
				v.SetDefault(Target, "file")
			}

			//创建配置文件并写入内容
			newConfStr := strings.Replace(defaultConf, MACRO+LoggingDirectory, v.GetString(LoggingDirectory), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+LoggingActiveFileName, v.GetString(LoggingActiveFileName), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+LoggingWriteOutput, fmt.Sprintf("%t", v.GetBool(LoggingWriteOutput)), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+RotationMaxLines, fmt.Sprintf("%d", v.GetInt(RotationMaxLines)), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+RotationMaxFileSizeBytes, fmt.Sprintf("%d", v.GetInt(RotationMaxFileSizeBytes)), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+FlushingTimeIntervalSecs, fmt.Sprintf("%d", v.GetInt(FlushingTimeIntervalSecs)), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+PrependValue, v.GetString(PrependValue), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+FileRenamePolicy, v.GetString(FileRenamePolicy), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+MaxAge, v.GetString(MaxAge), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+MaxCount, fmt.Sprintf("%d", v.GetInt(MaxCount)), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+Gzip, v.GetString(Gzip), -1)
			newConfStr = strings.Replace(newConfStr, MACRO+Target, v.GetString(Target), -1)

			err = ioutil.WriteFile(confpath, []byte(newConfStr), os.ModePerm)
			if err != nil {
				return nil, reloadChan, nil, err
			}
		}
	}

	// Read from env vars
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Validate
	if err := validateConfig(v); err != nil {
		return nil, reloadChan, nil, err
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			if err := validateConfig(v); err != nil {
				logger.Err(err.Error())
				return
			}
			reloadChan <- getConfigStruct(v)
		}
	})
	// return output writer by passing the viper instance
	outputWriter, err := GetOutputWriter(v, logger)
	if err != nil {
		return nil, reloadChan, nil, err
	}

	// return struct
	return getConfigStruct(v), reloadChan, outputWriter, nil
}

func setDefaults(v *viper.Viper) {
	if v.GetString(LoggingDirectory) == "" {
		v.SetDefault(LoggingDirectory, "log")
	}

	if v.GetString(LoggingActiveFileName) == "" {
		v.SetDefault(LoggingActiveFileName, "out.log")
	}

	if v.GetString(LoggingWriteOutput) != "true" {
		v.SetDefault(LoggingWriteOutput, false)
	}

	if v.GetInt(RotationMaxLines) < 1 {
		v.SetDefault(RotationMaxLines, 100000)
	}

	if v.GetInt(RotationMaxFileSizeBytes) < 1 {
		v.SetDefault(RotationMaxFileSizeBytes, 5000000)
	}

	if v.GetInt(FlushingTimeIntervalSecs) < 1 {
		v.SetDefault(FlushingTimeIntervalSecs, 5)
	}

	if v.GetString(PrependValue) == "" {
		v.SetDefault(PrependValue, "")
	}

	if v.GetString(FileRenamePolicy) != "timestamp" && v.GetString(FileRenamePolicy) != "serial" {
		v.SetDefault(FileRenamePolicy, "timestamp")
	}

	if v.GetString(MaxAge) == "" {
		v.SetDefault(MaxAge, "30d")
	}

	if v.GetInt(MaxCount) < 1 {
		v.SetDefault(MaxCount, 100)
	}

	if v.GetString(Gzip) != "true" {
		v.SetDefault(Gzip, false)
	}

	if v.GetString(Target) == "" {
		v.SetDefault(Target, "file")
	}
}

func validateConfig(v *viper.Viper) error {
	// Validate strings
	for _, key := range []string{
		LoggingDirectory,
		LoggingActiveFileName,
		LoggingWriteOutput,
		PrependValue,
		FileRenamePolicy,
		MaxAge,
		Target,
	} {
		// If a string or bool value got successfully converted to integer,
		// then its incorrect
		if _, err := strconv.Atoi(v.GetString(key)); err == nil {
			return &ConfigValueError{key}
		}

		// File rename policy has to be either timestamp or serial
		if key == FileRenamePolicy &&
			(v.GetString(key) != "timestamp" && v.GetString(key) != "serial") {
			return ErrInvalidFileRenamePolicy
		}
	}

	// Validate integers
	for _, key := range []string{
		RotationMaxLines,
		RotationMaxFileSizeBytes,
		FlushingTimeIntervalSecs,
		MaxCount,
	} {
		// If an integer value was a string, it would come as zero,
		// hence its invalid
		if v.GetInt(key) == 0 {
			return &ConfigValueError{key}
		}
	}

	// Validate MaxAge
	maxAge := v.GetString(MaxAge)
	unit := maxAge[len(maxAge)-1:]
	_, err := strconv.Atoi(maxAge[0 : len(maxAge)-1])
	if err != nil {
		return ErrInvalidMaxAge
	}

	if unit != "d" && unit != "h" {
		return ErrInvalidMaxAge
	}

	return nil
}

func getConfigStruct(v *viper.Viper) *Config {
	return &Config{
		DirName:                  v.GetString(LoggingDirectory),
		ActiveFileName:           v.GetString(LoggingActiveFileName),
		WriteOutput:              v.GetBool(LoggingWriteOutput),
		RotationMaxLines:         v.GetInt(RotationMaxLines),
		RotationMaxBytes:         uint64(v.GetInt64(RotationMaxFileSizeBytes)),
		FlushingTimeIntervalSecs: v.GetInt(FlushingTimeIntervalSecs),
		PrependValue:             v.GetString(PrependValue),
		FileRenamePolicy:         v.GetString(FileRenamePolicy),
		MaxAge:                   getMaxAgeValue(v.GetString(MaxAge)),
		MaxCount:                 v.GetInt(MaxCount),
		Gzip:                     v.GetBool(Gzip),
		Target:                   v.GetString(Target),
	}
}

func getMaxAgeValue(maxAge string) int64 {
	unit := maxAge[len(maxAge)-1:]
	magnitude, _ := strconv.Atoi(maxAge[0 : len(maxAge)-1])

	if unit == "d" {
		return int64(magnitude) * 24 * 60 * 60
	}
	return int64(magnitude) * 60 * 60
}
