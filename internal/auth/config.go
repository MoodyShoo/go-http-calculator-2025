package auth

import "os"

const SignatureEnv = "JWT_SECRET"

type Config struct {
	Signature []byte
}

func configFromEnv() *Config {
	conf := &Config{
		Signature: []byte("very_secret"),
	}

	if secret := os.Getenv(SignatureEnv); secret != "" {
		conf.Signature = []byte(secret)
	}

	return conf
}
