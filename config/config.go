package config

import (
	"io/ioutil"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

type AuthConfig struct {
	JWT struct {
		Secret          string `yaml:"secret"`
		ExpirationHours int    `yaml:"expiration_hours"`
	} `yaml:"jwt"`
	Admin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Role     string `yaml:"role"`
	} `yaml:"admin"`
}

var (
	authConfig *AuthConfig
)

// LoadAuthConfig 加载认证配置
func LoadAuthConfig(configPath string) (*AuthConfig, error) {
	if authConfig != nil {
		return authConfig, nil
	}

	filename := filepath.Join(configPath, "auth.yaml")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &AuthConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// 对管理员密码进行哈希处理
	if !isPasswordHashed(config.Admin.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.Admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		config.Admin.Password = string(hashedPassword)
	}

	authConfig = config
	return config, nil
}

// GetAuthConfig 获取认证配置
func GetAuthConfig() *AuthConfig {
	return authConfig
}

// isPasswordHashed 检查密码是否已经过哈希处理
func isPasswordHashed(password string) bool {
	// bcrypt哈希的长度为60
	return len(password) == 60 && password[0] == '$'
}
