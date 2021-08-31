# guild-golems

Go-based server for a fantasy-themed API game

## Features

- Basic account functionality
- - Claim Account: `POST` to `https://guildgolems.io/api/v0/users/{username}/claim`
- - - Cannot claim if username length < 0 or if already claimed
- - Get Account Info: `GET` from `https://guildgolems.io/api/v0/my/account` including `Bearer Token` in `Authorization` header
- Basic world info
- - Request info: `GET` to `https://guildgolems.io/api/v0/locations`

## Roadmap / TODO

- Convert from using username as uuid to using token as uuid (should simplify some sections and makes more sense to me)
- Improve endpoints for world info
- - Fix enum not serializing to and from json correctly in economy/ideology
- - Separate region, locale, poi data
- - Add locale and poi data to db
- - Add interactions for POIs
- - Perhaps re-organize structs folder/packages during this task
- **All Caught Up - Add More!**

## Build & Run

Create `secrets.env` file in the project root containing `GG_ACCESS_SECRET=<SOME_SECRET_ALPHA_NUMERICS>` e.g. `GG_ACCESS_SECRET=AJSHF38FJ93SL98`

Build with `go build`, start with `./guild-golems`. Alternatively, `go run main.go`

Listens on port `50242`

## Changelog

### v0.0.1

- Updated to go 1.17
- Implemented user account claiming (incl. save to db) and getting account info
- - Implemented `auth` package for handling tokens and authentication
- - Implemented secure `/api/v0/my` subrouter
- - Implemented `github.com/joho/godotenv` to load `.env` file with auth secret
- Implemented `https://guildgolems.io/api/v0/locations` endpoint to show loaded world data
- Fixed bug with region not correctly loading economy, connected regions, or locales
- Added `secrets.env` and added it to `.gitignore`
- Added `db.UpdateWorld`
- Added `guild-golems` to `.gitignore` - its a binary

### v0.0.0

- Initial router setup
- Initial db setup
- Initial json import setup

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

### Design

- https://api.spacetraders.io/
- (Private) https://docs.google.com/document/d/15d-nC5dpiH19LH1sbWiUOM5Pjgr_Cjop-t_Dmuu2Xtc/edit
- (Private) https://keep.google.com/u/0/#LIST/1AyAhsCulc79U76hQK60tpjy9RaC5uQ6MdjHDYKDGrn8CsEPV56mWNezvrTPRdGA_cCrc9Q
