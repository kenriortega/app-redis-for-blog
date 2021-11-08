app-ch:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/cacheapp/main.go

app-rl:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/ratelimit/main.go


# Run redis for local env
redis:
	podman run --name redis --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 quay.io/bitnami/redis:latest

redis-cli:
	podman exec -ti redis redis-cli