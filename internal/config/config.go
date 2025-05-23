package config

import "time"

var JWTSecret = []byte("super-secret-key")

var TokenTTL = time.Minute * 15
