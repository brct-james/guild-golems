# brct-game

Go-based server for a fantasy-themed alchemy game

## Features

- Nothing

### Endpoints

- `GET: api/v0/locations` returns entire world json from DB

## Roadmap / TODO

- [In-Progress] Port progress from guild-golems following idioms
- - Working on restoring the following:
- - - auth/token-validation
- - - handlers/general
- - - - Refactor locationsOverview in particular - its uncommented and updated but not been streamlined nor made to conform to idioms
- - - handlers/secure
- - - rdb (specifically, how handling user vs world CRUD - do I want/need wrappers for this?)
- - - responses/responses
- Make auth secret automatically generate if doesn't exist rather than relying on a variable

## Build & Run

Ensure resjon container is running on the correct port: `docker run -di -p 6381:6379 --name rejson_brct-game redislabs/rejson:latest`

For the first run, ensure `refreshAuthSecret` in `main.go` is true. Make sure to set this to false for second run.

Build and start with `go build; ./brct-game`. Alternatively, `go run main.go`

Listens on port `50235`

redis-cli via `redis-cli -p 6381`

`FLUSHDB` for each database (`select #`)

Recommend running with screen `screen -S gg`. If get detached, can forcibly detach the old ssh session and reattach with `screen -Dr gg`

## Changelog

### v0.0.1

- Added user db, world db
- - World db can be loaded from json
- Automatically generates access secret
- Added `GET: /api/v0/locations`

### v0.0.0

- Initialization

## Reference

### Technical

- https://github.com/RedisJSON/RedisJSON
- https://github.com/nitishm/go-rejson
- https://oss.redis.com/redisjson/commands/
- https://tutorialedge.net/golang/go-redis-tutorial/
- https://github.com/go-redis/redis
- https://tutorialedge.net/golang/parsing-json-with-golang/
- https://tutorialedge.net/golang/creating-restful-api-with-golang/
- https://github.com/joho/godotenv
- https://github.com/golang-jwt/jwt

### Design

- https://api.spacetraders.io/
- https://spacetraders.io/docs/guide
- (Private) https://docs.google.com/document/d/15d-nC5dpiH19LH1sbWiUOM5Pjgr_Cjop-t_Dmuu2Xtc/edit
- (Private) https://keep.google.com/u/0/#LIST/1AyAhsCulc79U76hQK60tpjy9RaC5uQ6MdjHDYKDGrn8CsEPV56mWNezvrTPRdGA_cCrc9Q
- https://spacetraders.io/docs/ship-design
