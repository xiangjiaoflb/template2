package jsonlog

// XXX: Move it to constants.go if needed
const (
	// config 以下是避免引入 funnel， win下无 syslog.Writer 库
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

	//LogLevel 日志级别
	LogLevel = "loglevel"
)
