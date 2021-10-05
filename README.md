# brct-game

Go-based server for a fantasy-themed alchemy game

## Features

- Register a username and get your public user info ~~as well as the secure `/my/account` endpoint~~

### Endpoints

- `GET: api/v0/locations` returns entire world json from DB
- `POST: /api/v0/users/{username}/claim` attempts to claim the specified username, returns the user data after creation, including token which users must save to access private routes
- `GET: /api/v0/users/{username}` returns the public user data
- `GET: /api/v0/my/account` returns the private user data (includes token)

### Response Codes

```golang
case -3:
  message = "[CRITICAL_JSON_MARSHAL_ERROR] Server error in responses.JSON, could not marshal JSON_Marshal_Error response! PLEASE contact developer."
case -2:
  message = "[JSON_Marshal_Error] Responses module encountered an error while marshaling response JSON. Please contact developer."
case -1:
  message = "[Unimplemented] Unimplemented Feature. You shouldn't be able to hit this on the live build... Please contact developer"
case 0:
  message = "[Generic_Failure] Contact developer"
case 1:
  message = "[Generic_Success] Request Successful"
case 2:
  message = "[Auth_Failure] Token was invalid or missing from request. Did you confirm sending the token as an authorization header?"
case 3:
  message = "[Username_Validation_Failure] Please ensure username conforms to requirements and account does not already exist!"
case 4:
  message = "[DB_Save_Failure] Failed to save to DB"
case 5:
  message = "[Generate_Token_Failure] Username passed initial validation but could not generate token, contact Admin."
case 6:
  message = "[WDB_Get_Failure] Could not get from world DB"
case 7:
  message = "[UDB_Get_Failure] Could not get from user DB"
case 8:
  message = "[JSON_Unmarshal_Error] Error while attempting to unmarshal JSON from DB"
case 9:
  message = "[No_WDB_Context] Could not get WDB context from middleware"
case 10:
  message = "[No_UDB_Context] Could not get UDB context from middleware"
case 11:
  message = "[No_AuthPair_Context] Failed to get AuthPair context from middleware"
case 12:
  message = "[User_Not_Found] User not found!"
default:
  message = "[Unexpected_Error] ResponseCode not in valid enum range! Contact developer"
```

## Roadmap

### In-Progress

- Nothing

### Planned: v0 MVP

- Ratelimiting
- - Per-IP hard limit, slightly higher than per-token limit
- - Per-IP route-specific hard limit for claiming usernames
- - Per-Token limit
- - Perhaps using [toolbooth](https://github.com/didip/tollbooth)
- Achievements system
- - Achievements like "first million" (hit 1,000,000 coins) or "ratelimited" (X calls in the last hour).
- - Tracked per person as well as for leaderboards
- - Provide incremental goals to grow into
- Caching service - generate leaderboards every 15/30/60 minutes and cache them, same for location info (esp. #golems at each location) and calls per hour
- `/` UI homepage with info on the game
- `/docs` documentation route
- `/api` information on each api version like basic live/down status (future - once multiple versions are live)
- `/api/v0` v0 api status as well as intresting metrics in the data field (totalCoins in circulation, totalCalls made to server, lastWipeTimestamp, avgServerTicksPerSecond in a rolling minute, etc.)
- `.../users` users summary (e.g. unique, active - call in last 5 min, etc.) example:

```json
{
  "uniqueUsers": ["Greenitthe", ...],
  "activeUsers": ["Greenitthe", ...],
  "usersWithAchievement": {
    "ratelimited": ["Greenitthe", ...],
    "first_million": ["Greenitthe", ...],
    ...
  }
}
```

- `.../leaderboard` info on top 10 players by various metrics, example:

```json
{
  "coinleaders": [
    {
      "name": "Greenitthe",
      "totalGolems": 47,
      "coins": 123456,
      "callsLastHour": 212,
      "achievementCount": 7,
      "creationTimestamp": 0000000000
    },
    ...
  ],
  "golemleaders": [{...}, ...],
  "achievementLeaders": [{...}, ...]
}
```

- `.../achievements` info on each achievement like name, description/criteria, list of players with it, etc.
- for v0 give each task its own endpoint, may re-evaluate this structure later
- - `.../my/harvesters` collecting free resources from nodes
- - - v0: simply collecting X resource at location taking Y time
- - `.../my/couriers` transporting materials between two locations
- - - v0: simply moving resources between locations, based on a set speed and capacity
- - `.../my/artisans` converting resources into products
- - - v0 will just see simple recipes, "make 2 C using 2 A and 3 B". In the future there will be more depth, perhaps with quality or yield implications to identifying the right ratio of resources (e.g. A+B always yields C between X and Y ratio, but has max yield at some particular ratio) and/or techniques or tools that change recipes and will be complicated with multiple steps for higher tier products
- - `.../my/merchants` buying/selling
- - - v0: simple buy and sell tasks at set market prices
- - - v0: basic fog of war, as using this route to get market listings, so merchants must be at a location to get the values of that route
- - add action routes beneath each of these task endpoints (e.g. move, buy, sell, collect, create, etc.)
- - `.../my/invokers` summon new golems for each given task and cast other spells
- - - v0: single golem type per task, takes mana to summon, summoner golems generate mana
- - - v0: spell to move a golem instantly between locations, can be used with a courier for instant moving of resources as well, mana cost by weight/volume
- rework `/locations` routes to be more specific, include list of users with golems at each location and how many
- `.../my/inventory` inventory report showing what resources are at each location
- `.../my/golems` golem report showing what golems are at each location and what task they are doing if active

### Planned: Unscheduled

- Performance monitoring to see what calls are expensive as well as what are most used to see where to focus optimization or streamlining
- `/ui` routes for each endpoint that work for players who don't want to use the api
- - `/ui` page with a guide/tutorial
- - `/ui/...` pages visualizing the returned data and/or providing interactive buttons and such
- Make auth secret automatically generate if doesn't exist rather than relying on a variable
- Event tagging: report a unique error/failure/event # in error message response for users to send for troubleshooting
- More full-featured fog of war - player 'home base' that must be communicated with using couriers/messengers/mana-using spells
- 'bandwidth' Resource used by calls, with more expensive calls costing more, to incentivize optimization. Replaces ratelimiting (leave a hard rate limit at some high value just in-case)
- World Events
- - Contribute resources and money towards world events that unlock new content once they are completed
- - - E.g. "crusade to clear the huge dungeon outside of town needs potions to supply its fighters, unlocking this area for gathering after", "archmage is investigating the creation of the philosopher stone, once complete everyone can buy a stone that unlocks new recipes", etc.
- complicate `harvesting`: some kind of 'harvesting techniques' perhaps, to add depth - say 10 different harvesting techniques for each type of harvesting, along with some randomness in yield, so that you need to track yield over time with each technique for each resource to optimize effectively
- complicate `couriers`: incorporate route risk and courier preparedness (e.g. more couriers for less resources make them faster and easier to defend, and some couriers may have a combat focus)
- complicate `artisans`:
- complicate `merchants`: more dynamic pricing that is player affected can be implemented, as well as some way to add depth to the task like setting up your own storefront
- complicate `summoning`: multiple tiers or more in-depth customization
- expanded achievements system - could have achievement points that can be spent on upgrades
- investigate using a job queue/dispatcher with one worker per user token handling any requests sent to that user's channel for distributed computing
- - global stuff would be handled by its own channel and worker
- - thus only one worker is ever going to be accessing any given portion of the DB at once

## Build & Run

Ensure resjon container is running on the correct port: `docker run -di -p 6381:6379 --name rejson_brct-game redislabs/rejson:latest`

For the first run, ensure `refreshAuthSecret` in `main.go` is true. Make sure to set this to false for second run.

Build and start with `go build; ./brct-game`. Alternatively, `go run main.go`

Listens on port `50235`

redis-cli via `redis-cli -p 6381`

`FLUSHDB` for each database (`select #`)

Recommend running with screen `screen -S brct-game`. If get detached, can forcibly detach the old ssh session and reattach with `screen -Dr brct-game`

## Changelog

### v0.0.2

- Nothing

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
- Renamed handlers/general to handlers/public
- Refactored handlers/secure
- Refactored auth/token-validation
- Added `GET: /api/v0/my/account`
- - Requires bearer token authorization header, provides full user info
- responses now return error instead of panicing if json prettification fails...
- error messages now include code literal meaning as first part of message

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
