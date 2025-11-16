package config

type Config struct {
	Env          string       `yaml:"env" env-default:"dev"`
	Host         string       `yaml:"host"`
	Port         int          `yaml:"port"`
	Minio        Minio        `yaml:"minio"`
	PresignedUrl PresignedUrl `yaml:"presigned_url"`
}

type Minio struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	UseSSL     bool   `yaml:"use_ssl"`
	BucketName string `yaml:"bucket"`
}

type PresignedUrl struct {
	ExpiryHours int `yaml:"expiry_hours"`
}
