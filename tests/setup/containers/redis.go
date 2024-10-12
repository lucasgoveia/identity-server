package containers

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"identity-server/config"
)

func SetupTestRedis() (*config.RedisConfig, func(), error) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"docker.io/redis",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)

	if err != nil {
		return nil, nil, err
	}

	teardown := func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			log.Errorf("failed to terminate container: %s", err)
		}
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, teardown, err
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, teardown, err
	}

	redisUrl := fmt.Sprintf("%s:%s", host, port)

	return &config.RedisConfig{
		Url:      redisUrl,
		Password: "",
		Username: "",
	}, teardown, nil
}
