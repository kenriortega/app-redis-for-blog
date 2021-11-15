app-ch:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/cacheapp/main.go

app-rl:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/ratelimit/main.go

app-ms:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=web
app-ms-seed:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=seed
app-ms-indices:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=create:index
# Run redis for local env
redis:
	podman run --name redis --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 quay.io/bitnami/redis:latest
redis-mod:
	podman run --name redis-mod --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 localhost/redislab/redismod:latest

redis-cli:
	podman exec -ti redis-mod redis-cli