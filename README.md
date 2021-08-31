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

### v0.0.0

- Initial router setup
- Initial db setup
- Initial json import setup

## API Thoughts

Consider best layout for location data: singular json file vs joining files for regions, locales, pois, and entities

/api/v0/locations/astria/_locale_/_poi_
See below for info on locales and pois. v0 only has one region, Astria. Entities are not implemented, instead all functionality is done via the poi itself (e.g. getting quests, conducting rituals, buying and selling). On the backend this information is basically a json lookup until you get to the interaction. Once multiplayer is implemented, will need to be a database lookup as player guild members, guild buildings, etc. will change.

ideology is based on the following axes:
economic freedom doctrine: Unregulated, Semi-Regulated, Regulated, Semi-Planned, Planned
social doctrine: Collectivist, Progressive, Centrist, Conservative, Individualist
state authority doctrine: Anarchic, Libertarian, Moderate, Controlling, Authoritarian

/api/v1/locations/_region_/_locale_/_poi_/_entity_

Regions are areas controlled by certain factions, like kingdoms, nations, etc.
GET: information

Locales are things like settlements, war camps, dungeons, forests, etc. that contain POIs
GET: information

POIs are things like taverns, markets, throne rooms, guild buildings, dungeons, etc.
additional note: some POIs must be discovered, and so also hold a discovered flag that must be true to be returned
GET: information

Entities are things like NPCs, enemies, important objects, etc.
GET entity:
get name, symbol, POI, overview/description/flavor, schedule
additional note: some entities have schedules and are only available in certain locations during certain times
POST: interact (create: e.g. take out new loan, summon new golem, etc.)
PUT: interact (edit: e.g. pay off loan, change market orders, etc.)

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
