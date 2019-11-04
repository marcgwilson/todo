package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Database string
	Port     int
	Limit    int64
}

func (r *Config) Addr() string {
	return fmt.Sprintf(":%d", r.Port)
}

func LookupConfig() (*Config, error) {
	var err error

	database := "todo.db"
	port := 8000
	limit := int64(20)

	if env, ok := os.LookupEnv("TODO_DB"); ok {
		database = env
	}

	if env, ok := os.LookupEnv("TODO_PORT"); ok {
		if port, err = strconv.Atoi(env); err != nil {
			return nil, fmt.Errorf("Error parsing TODO_PORT: %s", env)
		}
	}

	if env, ok := os.LookupEnv("TODO_LIMIT"); ok {
		if limit, err = strconv.ParseInt(env, 10, 64); err != nil {
			return nil, fmt.Errorf("Error parsing TODO_LIMIT: %s", env)
		}
	}

	return &Config{database, port, limit}, nil
}
