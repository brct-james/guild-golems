# guild-golems

Go-based server for a fantasy-themed API game

## Features

## Roadmap / TODO

## Build & Run

Build with `go build`, start with `./guild-golems`. Alternatively, `go run main.go`

Listens on port `50242`

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
