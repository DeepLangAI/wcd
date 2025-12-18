package conflib

type Logger struct {
	LogPath  string `yaml:"log_path"`  // 日志文件路径
	LogLevel string `yaml:"log_level"` // 日志级别
}
