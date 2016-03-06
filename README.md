# diglet

[![Build Status](https://travis-ci.org/buckhx/diglet.svg?branch=master)](https://travis-ci.org/buckhx/diglet)

A simple tool for solving common geospatial workflows.

* wms: Web Mapping Server (tile server)
* mbt: Tile builder for vector tiles into the mbtiles spec

# wms

A real-time tile server in a single binary

Here are some neat things that diglet wms does

* Backend changes are pushed to the front end in real time
  * Currently changes are registered from the kernel (inotify/kqueue/ReadDirectoryChangesW)
* Sniffs the tile format 
  * (pbf, json, gz, jpg, png, etc...)
* Source specific hooks in the works
  * (on PostGIS insert -> build mbtiles)
* HTTP/JSON-RPC/WS endpoints
* All packaged up in an itty-bitty binary

## Usage

```
NAME:
   diglet wms - Starts the diglet Web Map Service

USAGE:
   diglet wms [command options] mbtiles_directory

DESCRIPTION:
   Starts the diglet Web Map Service

OPTIONS:
   --port "8080"		Port to bind
   --cert, --tls-certificate 	Path to .pem TLS Certificate. Both cert & key required to serve HTTPS
   --key, --tls-private-key 	Path to .pem TLS Private Key. Both cert & key required to serve HTTPS
   --tms-origin			NOT IMPLEMENTED: Use TMS origin, SW origin w/ Y increasing North-wise

```

If --cert and --key are both set, content will be served over TLS (HTTPS) and unecrypted HTTP will return nothing or a
TLS error response

## Methods

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

#mbt

More info to come, but here's the current help message

```
NAME:
   diglet mbt - Builds an mbtiles database from the input data source

USAGE:
   diglet mbt [command options] input_source

DESCRIPTION:
   Builds an mbtiles database from the given format

OPTIONS:
   -o, --output 					REQUIRED: Path to write mbtiles to
   --input-type "sniff"					Type of input files, 'sniff' will pick type based on the extension
   -f, --force						Remove the existing .mbtiles file before running.
   -u, --upsert						Upsert into mbtiles instead of replacing.
   --layer-name "features"				Name of the layer for the features to be added to
   --desc, --description "Generated from Diglet"	Value inserted into the description entry of the mbtiles
   --extent "4096"					Extent of tiles to be built. Default is 4096
   --max, --max-zoom "10"				Maximum zoom level to build tiles for. Not Implemented.
   --min, --min-zoom "5"				Minimum zoom level to build tiles from. Not Implemented.
   --filter 						Only include fields keys in this comma delimited list.	EXAMPLE --filter name,date,case_number,id	NOTE all fields are lowercased and non-word chars replaced with '_'
   --csv-lat "latitude"					
   --csv-lon "longitude"				
   --csv-geometry "geometry"				Column containing geometry in geojson-like 'coordinates' form
   --csv-delimiter ","	
```

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

## Licenses

Here are dependencies and their licenese

* buckhx/diglet: MIT
* buckhx/mbtiles: MIT
* codegangsta/cli: MIT
* qedus/osmpbf: MIT
* gorilla: BSD 3-Clause
* mattn/go-sqlite3: MIT
* go-std-lib: BSD 3-Clause
