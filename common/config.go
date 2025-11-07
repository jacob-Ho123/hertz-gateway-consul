package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/yaml.v3"
)

const (
	ReviewingBlockingKey = "reviewing_blocking.yaml"
)

type CommonConfig struct {
	OTLPEndpoint string `json:"otel_endpoint" yaml:"otel_endpoint"`
	MetricsAddr  string `json:"metrics_addr" yaml:"metrics_addr"`
	MetricsPath  string `json:"metrics_path" yaml:"metrics_path"`
	ConsulAddr   string `json:"consul_addr" yaml:"consul_addr" mapstructure:"consul_addr"`
	ConsulDc     string `yaml:"consul_dc" mapstructure:"consul_dc"`
}

type AppMappingConfig struct {
	AppId         string `json:"app_id" yaml:"app_id"`
	ReqMapConfig  string `json:"req_map_config" yaml:"req_map_config"`
	RespMapConfig string `json:"resp_map_config" yaml:"resp_map_config"`
}

type ProxyConfig struct {
	Common            CommonConfig       `json:"common" yaml:"common"`
	AppMappingConfigs []AppMappingConfig `json:"app_mapping" yaml:"app_mapping"`
	ReqReplacerMap    map[string]*strings.Replacer
	RespReplacerMap   map[string]*strings.Replacer
}

func LoadConfig(filename string) (*ProxyConfig, error) {
	hlog.Infof("Loading default config, file: %s", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var defaultConfig ProxyConfig
	ext := filepath.Ext(filename)

	switch ext {
	case ".json":
		err = json.Unmarshal(data, &defaultConfig)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &defaultConfig)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		return nil, err
	}
	defaultConfig.ReqReplacerMap = make(map[string]*strings.Replacer)
	defaultConfig.RespReplacerMap = make(map[string]*strings.Replacer)
	hlog.Infof("defaultConfig AppMappingConfigs: %+v", defaultConfig.AppMappingConfigs)
	for _, appMappingConfig := range defaultConfig.AppMappingConfigs {
		reqMap := make(map[string]string)
		_err := json.Unmarshal([]byte(appMappingConfig.ReqMapConfig), &reqMap)
		if _err == nil {
			var reqReplaceParam []string
			for k, v := range reqMap {
				if strings.TrimSpace(v) != "" && strings.TrimSpace(k) != "" {
					reqReplaceParam = append(reqReplaceParam, "\""+strings.TrimSpace(v)+"\":", "\""+strings.TrimSpace(k)+"\":")
				}
			}
			if len(reqReplaceParam) > 0 {
				hlog.Infof("AppId: %s, reqReplaceParam: %+v", appMappingConfig.AppId, reqReplaceParam)
				defaultConfig.ReqReplacerMap[strings.TrimSpace(appMappingConfig.AppId)] = strings.NewReplacer(reqReplaceParam...)
			}
		}
		respMap := make(map[string]string)
		_err = json.Unmarshal([]byte(appMappingConfig.RespMapConfig), &respMap)
		if _err == nil {
			var respReplaceParam []string
			for k, v := range respMap {
				if strings.TrimSpace(v) != "" && strings.TrimSpace(k) != "" {
					respReplaceParam = append(respReplaceParam, "\""+strings.TrimSpace(k)+"\":", "\""+strings.TrimSpace(v)+"\":")
				}
			}
			if len(respReplaceParam) > 0 {
				hlog.Infof("AppId: %s, respReplaceParam: %+v", appMappingConfig.AppId, respReplaceParam)
				defaultConfig.RespReplacerMap[strings.TrimSpace(appMappingConfig.AppId)] = strings.NewReplacer(respReplaceParam...)
			}
		}
	}
	return &defaultConfig, nil
}
