# guild-golems

Go-based server for a fantasy-themed guild management game

## Features

- Basic account functionality
- - Claim Account: `POST: https://guildgolems.io/api/v0/users/{username}/claim`
- - - Don't forget to save the token from the response to your claim request. You must use this as a bearer token in the auth header for secure `/my/` routes
- - - Must include only letters, numbers, `-`, and `_`.
- - Get public user info at `/api/v0/users/{username}` and get private user info including token at `/api/v0/my/account`
- Basic location info
- - Get world json: `GET: https://guildgolems.io/api/v0/locations`
- Summon Golems using Mana
- - `invokers` amplify your mana regen
- - Mana regen is calculated every time `secureGetUser` is called
- - `harvesters` gather resources from nodes in the world
- Have golems travel between locations
- Leaderboards based on various criteria
- - Note: At present leaderboards are not being generated nor cached, so no rankings are returned
- Get lists of all unique, active, etc. users

---

### Endpoints

- `GET: /api/v0/leaderboards` list all available leaderboards and their descriptions
- `GET: /api/v0/leaderboards/{board}` get the specified leaderboard rankings
- `GET: /api/v0/locations` returns entire world json from DB
- `GET: /api/v0/users` returns lists of registered usernames with various filters: unique, active, etc.
- `GET: /api/v0/users/{username}` returns the public user data
- `POST: /api/v0/users/{username}/claim` attempts to claim the specified username, returns the user data after creation, including token which users must save to access private routes
- `GET: /api/v0/my/account` returns the private user data (includes token)
- `GET: /api/v0/my/golems` list all golems owned
- `GET: /api/v0/my/golems/{archetype}` list all golems owned filtered by archetype
- `GET: /api/v0/my/golem/{symbol}` get info on the specified golem
- `PUT: /api/v0/my/golem/{symbol}` change golem task/status based on request body (see requests section below)
- `GET: /api/v0/my/rituals` list all known rituals
- `GET: /api/v0/my/rituals/{ritual}` show information on a particular ritual
- `POST: /api/v0/my/rituals/{ritual}` attempt to do the given ritual
- - `summon-invoker` Spend mana to summon a new invoker, who can be used to help generate even more mana.
- - `summon-harvester` Spend mana to summon a new harvester, who can be used to gather resources from nodes in the world.

---

### Requests

- `PUT: /api/v0/my/golem/{symbol}` expects the following body:

```json
{
    "new_status":"",
    "instructions": {...}
}
```

- - Where new_status is the desired task from the set [`idle`, `harvesting`, `traveling`, `invoking`]
- - Where instructions contain key:value pairs specific to each type of activity
- - - `idle` instructions | {}
- - - `traveling` instructions | {"route": "A-G|A-SWF|WALK"}

---

### Response Codes

See `responses.go`

## Roadmap

---

### Ongoing

- All routes should use kebab-case
- All json & code should use snake_case

---

### In-Progress

**[v0.0.3]** Harvesters v0

- Harvesters collecting free resources from nodes
- - v0: simply collecting X resource at location taking Y time

**[v0.0.3]** Game Variables Refactor

- Game variables should be defined in gamelogic, for example, golem capacity `gamelogic.InvokerCapacity`
- - Refactor and ensure this is occurring throughout the code

### Planned: v0.0.5 MVP

**[v0.0.4]** Inventories v0

- `.../my/inventory` inventory report showing what resources are at each location
- `.../my/couriers` transporting materials between two locations
- - v0: simply moving resources between locations, based on a set speed and capacity
- - how will inventories work? one thought is giving each golem an inventory, another is having a localized inventory for each area per-player. most likely want to harvest into a local inventory, moved between locales by loading and unloading from golem inventories

**[v0.0.5]** Merchants v0

- `.../my/merchants` buying/selling
- - v0: simple buy and sell tasks at set market prices

---

### Planned: v0.1.0 MVP

**[v0.0.6]** Rituals v0

- cast spells
- - v0: spell to move a golem instantly between locations, can be used with a courier for instant moving of resources as well, mana cost by weight/volume
- `rituals` should be loaded from json into memory, User.KnownRituals should instead store the symbols and the ListRituals handler should lookup the rituals from the json map

**[v0.0.7]** Various Refactors

- Refactor the bloat in schema & handlers (helper funcs) into more appropriate locations
- Convert all routes to kebab-case, all json and private vars/funcs to snake_case

**[v0.0.7]** Various Endpoints & Merchants FOW

- `/api` information on each api version like basic live/down status (future - once multiple versions are live)
- `/api/v0` v0 api status as well as intresting metrics in the data field (totalCoins in circulation, totalCalls made to server, lastWipeTimestamp, avgServerTicksPerSecond in a rolling minute, etc.)
- `.../my/golems` golem report showing what golems are at each location and what task they are doing if active
- `.../my/merchants`
- - v0: basic fog of war, as using this route to get market listings, so merchants must be at a location to get the values of that route

**[v0.0.8]** Achievements v0

- Achievements like "first million" (hit 1,000,000 coins) or "ratelimited" (X calls in the last hour).
- Tracked per person as well as for leaderboards
- Provide incremental goals to grow into
- `.../achievements` info on each achievement like name, description/criteria, list of players with it, etc.

**[v0.0.8]** Caching/Metrics v0

- generate leaderboards every 15/30/60 minutes and cache them, same for location info (esp. #golems at each location) and calls per hour
- rework `/locations` routes to be more specific, include list of users with golems at each location and how many
- Calculate and cache leaderboards in metric-db
- Tracking all users ever registered, persistent through wipes in metric-db

**[v0.0.9]** Ratelimiting v0

- Per-IP hard limit, slightly higher than per-token limit
- Per-IP route-specific hard limit for claiming usernames
- Per-Token limit
- Perhaps using [toolbooth](https://github.com/didip/tollbooth)

**[v0.1.0]** Site UI v0

- `/` UI homepage with info on the game
- `/docs` documentation route

---

### Planned: v0.1.5 MVP

**[v0.1.1]** Artisans v0

- `.../my/artisans` converting resources into products
- - v0 will just see simple recipes, "make 2 C using 2 A and 3 B". In the future there will be more depth, perhaps with quality or yield implications to identifying the right ratio of resources (e.g. A+B always yields C between X and Y ratio, but has max yield at some particular ratio) and/or techniques or tools that change recipes and will be complicated with multiple steps for higher tier products

**[v0.1.2]** Engineers v0

- `.../my/engineers` building structures in the world
- - perhaps each location's central inventory can be upgraded, could construct warehouses at each location to hold more
- - tiers of golems and of buildings
- - pull resources while in 'building' status from the local inventory
- - more engineers present build faster

**[v0.1.3]** Resource Node Quantities and Replenishment

- Add node quantities to resource nodes, along with replenishment
- - Figure out a good way to balance this with race conditions when multiple users harvest at once, and replenishment having an unknown number of people harvesting at once

---

### Planned: Unscheduled

- semi-secure routes which will display more information if authorized (for fog of war on location routes, for example) - or is this being handled by using a separate route for markets and things?
- Define gamevars and other settings like those in main.go in config files rather than code
- Rituals v1
- - Specify a location for the ritual
- - Specify a number of repetitions for the ritual (batch calling)
- - Revisit invokers, rituals should require X amount of invokers at the location to work
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
- complicate golems by adding energy. Golems have energy which goes down after each task and slowly regenerates.
- complicate `harvesting`: some kind of 'harvesting techniques' perhaps, to add depth - say 10 different harvesting techniques for each type of harvesting, along with some randomness in yield, so that you need to track yield over time with each technique for each resource to optimize effectively
- complicate `couriers`: incorporate route risk and courier preparedness (e.g. more couriers for less resources make them faster and easier to defend, and some couriers may have a combat focus)
- complicate `artisans`:
- complicate `merchants`: more dynamic pricing that is player affected can be implemented, as well as some way to add depth to the task like setting up your own storefront
- complicate `summoning`: multiple tiers or more in-depth customization
- expanded achievements system - could have achievement points that can be spent on upgrades
- investigate using a job queue/dispatcher with one worker per user token handling any requests sent to that user's channel for distributed computing
- - global stuff would be handled by its own channel and worker
- - thus only one worker is ever going to be accessing any given portion of the DB at once
- rather than just calling CalculateManaRegen in secureGetUser, use a GetMana function that is only called when absolutely necessary - right now these are identical solutions but in the future other things will use secureGetUser that have nothing to do with mana
- - Also would need a flushUserUpdates or something like that for things like `/my/account` which need everything up to date
- Productionalize parsing json body: [link](https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body)

---

## Build & Run

Ensure resjon container is running on the correct port: `docker run -di -p 6380:6379 --name rejson_guild-golems redislabs/rejson:latest`

For the first run, ensure `refreshAuthSecret` in `main.go` is true. Make sure to set this to false for second run.

Build and start with `go build; ./guild-golems`. Alternatively, `go run main.go`

Listens on port `50242`

redis-cli via `redis-cli -p 6380`

`FLUSHDB` for each database (`select #`)

Recommend running with screen `screen -S guild-golems`. If get detached, can forcibly detach the old ssh session and reattach with `screen -Dr guild-golems`

---

## Changelog

### v0.0.3

- Refactored ChangeGolemTask into smaller functions

---

### v0.0.2

- Golems v0
- - Golems are created via rituals
- - `GET .../my/golems` list golems
- - - `GET.../my/golems/{archetype}` list golems filtered by archetype
- - `../my/golem/{symbol}` get info on and manage individual golems
- - - `GET` gives info on the specified invoker
- - - `PUT` allows changing the golem status to a new task
- Rituals v0
- - `GET .../my/rituals` list rituals
- - - `POST .../my/rituals/summon-invoker` create new invoker
- - - `GET .../my/rituals/summon-invoker` information on the invoker summoning ritual
- - - `POST .../my/rituals/summon-harvester` create new harvester
- - - `GET .../my/rituals/summon-harvester` information on the harvester summoning ritual
- - User holds list of ritual keys rather than list of rituals themselves
- Implemented the basics of the mana system and regeneration
- Changing golem status
- - Add travel time calculation
- Mana calculation only counts invokers with the 'invoking' status
- Added leaderboards endpoints (not being generated and cached yet)
- - `GET .../leaderboards` list leaderboards
- - `GET .../leaderboards/{board}` get leaderboard rankings
- Added users endpoint and lists of unique, active, etc. users
- - `GET .../users`
- - Implemented `metrics` package to handle these and other metrics
- Split timecalc functions into their own package
- Refactored world into multiple parts, e.g. world, regions, locales as separate json & database entries
- - Linked by symbols, no hierarchy
- - Json stored in individual files, loaded into `regions` `locales` etc maps e.g. `map[string]etc` where the string is Symbol for DB
- - - I have split json out and updated schema files, updated how the world is loaded from json into the DB, created some helper functions in the world schema for accessing the data, and updated everything that uses the WDB with the new functions

---

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

---

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
- https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql

### Design

- https://api.spacetraders.io/
- https://spacetraders.io/docs/guide
- (Private) https://docs.google.com/document/d/15d-nC5dpiH19LH1sbWiUOM5Pjgr_Cjop-t_Dmuu2Xtc/edit
- (Private) https://keep.google.com/u/0/#LIST/1AyAhsCulc79U76hQK60tpjy9RaC5uQ6MdjHDYKDGrn8CsEPV56mWNezvrTPRdGA_cCrc9Q
- https://spacetraders.io/docs/ship-design
