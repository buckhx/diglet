# diglet

[![Build Status](https://travis-ci.org/buckhx/diglet.svg?branch=master)](https://travis-ci.org/buckhx/diglet)

A real-time tile server in a single binary

Here are some neat things that diglet does

* Backend changes are pushed to the front end in real time
  * Currently changes are registered from the kernel (inotify/kqueue/ReadDirectoryChangesW)
* Sniffs the tile format 
  * (pbf, json, gz, jpg, png, etc...)
* Source specific hooks in the works
  * (on PostGIS insert -> build mbtiles)
* HTTP/JSON-RPC/WS endpoints
* All packaged up in an itty-bitty binary

# Usage
 
    diglet start --port 80 --mbtiles ~/path/to/dir/with/mbtiles_files/

--mbtiles: Path to local directory containing mbtiles files. NOTE only serves files
with .mbtiles extension

--port: default is 8080

# Methods

The following methods are available via the HTTP API. The other methods in the [app definition](diglet/app.go) are for use with WS
endpoints and mainly deal with tile subscriptions, which will remain undocumented until an official client is released
for them. The parameter {tileset-slug} refers to the mbtiles file on disk witout the extenstion and the name .

###ListTilesets

    GET /tileset/

List the .mbtiles on disk from --mbtiles and their attributes. The keys are the tileset-slug of the tilesets.

###GetTileset

    GET /tileset/{tileset-slug}

Get information about the specific tileset. This information is populated from the mbtiles metadata table.

###GetRawTile

    GET /tileset/{tileset-slug}/{z}/{x}/{y}

Get the tile at the given coordinates and return the contents as the response body. 
Passing json=true as a will return the tile as a json object with it's coordinates 

## Releases

Diglet uses a go library for reading mbtiles that depends on some C code, this makes
crosscompiling a bit of a pain, so only linux binaries will be published until there
is a need for more platforms. If a major release is it, diglet will be published for
most platforms.

We'll include a script and or directions for building on your own.

#### Minor releases will be built for linux x64 and i386

#### Major releases will have platform specific binaries

## Some food for thought

##### A tile is a girls best friend
##### A tile is forever
##### Divas. Desire. Tiles.
##### Our reputation shines as brightly as our tiles
##### The ultimate in luxury and tiles
##### Tiles by the Yard
##### The Tiler of Kings
##### Every tile begins with z
##### Your tiles should be as unique as you are
##### Tiles of brillance
##### Tiles, together.
##### Always the crowning tile
##### Strikingly brilliant. Remarkable. Tiles.
##### Preciously exclusive tiles
##### Tiles that brings out the luxury in you

