package config

type Config struct {
	Env          string       `yaml:"env" env-default:"dev"`
	Host         string       `yaml:"host"`
	Port         int          `yaml:"port"`
	Minio        Minio        `yaml:"minio"`
	PresignedUrl PresignedUrl `yaml:"presigned_url"`
}

type Minio struct {
	Endpoint       string `yaml:"endpoint"`
	PublicEndpoint string `yaml:"public_endpoint"`
	AccessKey      string `yaml:"access_key" env:"MINIO_ROOT_USER"`
	SecretKey      string `yaml:"secret_key" env:"MINIO_ROOT_PASSWORD"`
	UseSSL         bool   `yaml:"use_ssl" env:"MINIO_USE_SSL"`
	BucketName     string `yaml:"bucket" env:"MINIO_BUCKET_NAME"`
}

type PresignedUrl struct {
	ExpiryHours int `yaml:"expiry_hours"`
}
