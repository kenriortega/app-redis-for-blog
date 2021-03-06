app-ch:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/cacheapp/main.go

app-rl:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/ratelimit/main.go

app-ms:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=web
app-ms-seed-hash:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=seed:hash
app-ms-seed-json:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=seed:json
app-ms-indices:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=create:index
app-ms-indices-json:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minisearch/main.go --action=create:index:json

app-bk-producer:
	MQTT_TOPIC='/hfp/v2/journey/+/vp/bus/#' MQTT_BROKER='mqtt.hsl.fi' MQTT_PORT=8883 MQTT_N_WORKERS=2 REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minibroker/producer/main.go
app-bk-consumer:
	REDIS_URI="localhost:6379" REDIS_PASS="" go run examples/minibroker/consumer/main.go
# Run redis for local env
redis:
	podman run --name redis --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 quay.io/bitnami/redis:latest
redis-mod:
	podman run --name redis-mod --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 localhost/redislabs/redismod:latest

redis-cli:
	podman exec -ti redis-mod redis-cli
