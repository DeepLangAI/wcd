package conf

import (
	"fmt"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"os"
	"path/filepath"

	conflib "github.com/DeepLangAI/go_lib/conf"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Mongo           *Mongo         `yaml:"mongo"`
	MongoLingowhale *Mongo         `yaml:"mongo_lingowhale"`
	Logger          conflib.Logger `yaml:"logger"`
	Server          Server         `yaml:"server"`
	ApiDomain       ApiDomain      `yaml:"api_domain"`
	Parse           Parse          `yaml:"parse"`
}

type Parse struct {
	Label Label `yaml:"label"`
	Crawl Crawl `yaml:"crawl"`
}

type Label struct {
	UseMock bool `yaml:"use_mock"`
}

type Crawl struct {
	HtmlCacheHours int `yaml:"html_cache_hours"`
}

type ApiDomain struct {
	AiApi      string `yaml:"ai_api"`
	CrawlerApi string `yaml:"crawler_api"`
}

type Server struct {
	Port string `yaml:"port"` // 服务端口
	Name string `yaml:"name"`
}

type Mongo struct {
	Addr   string `yaml:"addr"`
	DbName string `yaml:"db_name"`
}

var (
	ConfigData Config
)

func GetConfig() Config {
	return ConfigData
}

func Init() {
	env := os.Getenv(constslib.ModeEnvName)
	if env == "" {
		env = constslib.ModeEnvDev
	}
	filePath := fmt.Sprintf(filepath.Join(GetProjectPath(), "./conf/config_%s.yaml"), env)
	fmt.Println("配置文件：" + filePath)
	dataBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败：%s", err))
	}

	err = yaml.Unmarshal(dataBytes, &ConfigData)
	if err != nil {
		panic(fmt.Sprintf("解析配置文件失败：%s", err))
	}
}

var projPath = ""

func GetProjectPath() string {
	if projPath != "" {
		return projPath
	}
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 从当前工作目录向上遍历，寻找main.go文件
	for {
		info, err := os.Stat(filepath.Join(cwd, "main.go"))
		if err == nil && !info.IsDir() {
			// 找到main.go，返回当前目录作为项目路径
			return cwd
		} else if !os.IsNotExist(err) {
			// 其他错误
			return ""
		}

		// 如果没找到，尝试进入上一级目录
		cwd = filepath.Dir(cwd)
		if cwd == "/" || cwd == "" {
			// 如果到达根目录仍然没找到，返回错误
			return ""
		}
	}

	// 不应达到这里，但为了编译器的满意度，返回一个空字符串
	return ""
}
