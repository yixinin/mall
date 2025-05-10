package main

type Config struct {
	Jwt  Jwt   `yaml:"jwt"`
	Mini WxApp `yaml:"mini"`
	Open WxApp `yaml:"open"`
}

type WxApp struct {
	AppID  string `yaml:"appid"`
	Secret string `yaml:"secret"`
}

type Jwt struct {
	Secret string `yaml:"secret"`
}
