# brct-game

Go-based server for a fantasy-themed alchemy game

## Features

- Register a username and get your public user info ~~as well as the secure `/my/account` endpoint~~

### Endpoints

- `GET: api/v0/locations` returns entire world json from DB

## Roadmap

### In-Progress

- Port progress from guild-golems following idioms
- - Working on restoring the following:
- - - auth/token-validation
- - - handlers/secure

### Planned: Next Update

- responses should return error instead of panicing if json prettification fails... will involve refactoring anything using `responses.JSON()`, `responses.FormatResponse()`, or `responses.SendRes()` to handle the 2nd return

### Planned: Unscheduled

- Make auth secret automatically generate if doesn't exist rather than relying on a variable
- Error tagging: report a unique error/failure # in error message for users to send for troubleshooting

## Build & Run

Ensure resjon container is running on the correct port: `docker run -di -p 6381:6379 --name rejson_brct-game redislabs/rejson:latest`

For the first run, ensure `refreshAuthSecret` in `main.go` is true. Make sure to set this to false for second run.

Build and start with `go build; ./brct-game`. Alternatively, `go run main.go`

Listens on port `50235`

redis-cli via `redis-cli -p 6381`

`FLUSHDB` for each database (`select #`)

Recommend running with screen `screen -S brct-game`. If get detached, can forcibly detach the old ssh session and reattach with `screen -Dr brct-game`

## Changelog

### v0.0.1

- Added user db, world db
- - World db can be loaded from json
- Automatically generates access secret
- Added `GET: /api/v0/locations`
- - Refactored locationsOverview
- Added `GET: /api/v0/users/{username}`
- - Refactored usernameInfo
- Added `POST: /api/v0/users/{username}/claim`
- - Refactored usernameClaim
- All other holdover general endpoints have had their placeholder info refactored
- Added user schema
- - user.go defines not only the `User` struct but also the `PublicUserInfo` struct as well as `NewUser()`, `CheckForExistingUser()`, and `GetUserFromDB()` funcs
- - - I decided to put these in the schema files for now, as that makes the most sense IMO - rdb is just for interacting with the DB, it shouldn't have anything to do with the data

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
