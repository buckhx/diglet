# diglet

[![Build Status](https://travis-ci.org/buckhx/diglet.svg?branch=master)](https://travis-ci.org/buckhx/diglet)
[![Go Report Card](https://goreportcard.com/badge/github.com/buckhx/diglet)](https://goreportcard.com/report/github.com/buckhx/diglet)

A simple tool for solving common geospatial workflows.

* [wms](#wms): Web Mapping Server (tile server)
* [mbt](#mbt): Tile builder for vector tiles into the mbtiles spec
* fence: A geofence utility
* dig: A geocoder based on OSM data

This project is under heavy development and APIs/CLIs are subject to change until further notice.

## Installation

Install diglet into /usr/local/bin (assuming you have permissions to) like so

    curl -sSL https://raw.githubusercontent.com/buckhx/diglet/master/scripts/install.py | sudo python - /usr/local/bin
    
If you don't have permission at /usr/local/bin, try something like this where you extend your PATH

    BINDIR=~/diglet
    mkdir -p $BINDIR
    export PATH=$PATH:$BINDIR
    curl -sSL https://raw.githubusercontent.com/buckhx/denv/master/scripts/install.py | python - $BINDIR

If you just want the binary, the final arg to the install py is a directory to download into

Or do it manually by going to the releases page and download the diglet artifact https://github.com/buckhx/diglet/releases/latest

Currently only building 64 bit linux for simplicity, but will build for more archs as things stabilize.
Instructions for [building](#building)

# wms

A real-time tile server in a single binary

Here are some neat things that diglet wms does

* Read-through mmap LRU cache
* Easy peasy-lemon squeezy HTTPS
* Sniffs the tile format and set Content-Type: (pbf, json, gz, jpg, png, etc...)
* HTTP/JSON-RPC/WS endpoints

Some things in the works

* Backend changes are pushed to the front end in real time (via websockeys)
* Source specific hooks (on PostGIS insert -> build mbtiles)

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

###Gallery

    GET /tileset/gallery/{tileset-slug}

A simple gallery to view your tiles with. Only supports vector tiles for now. ?lat={}&lon={}&zoom{} will zoom to desired location.

#mbt

Will build tiles from either geojson or a csv.

If a csv is used, either --csv-lat/--csv-lon or --csv shape must be set to read the coordinates correctly.
The csv-shape is a list of list of [[],[],[]] and will only render a single, exterior ring polygon per line.
Csv also requires a named header.

Geojson is fair-game, no support for GeometryCollection or Topojson

--filter if included will only include these columns in this properties. Includes all if not added

Valid extentsions for mbtiles are .mbtiles or .mbt

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
   --max, --max-zoom "10"				Maximum zoom level to build tiles for
   --min, --min-zoom "5"				Minimum zoom level to build tiles from
   --filter 						Only include fields keys in this comma delimited list
   --csv-lat						Column containing latitude					
   --csv-lon						Column containint longitude
   --csv-shape						Column containing geometry in geojson-like 'coordinates' form
   --csv-delimiter ","	
```

#fence & dig

I did a geofencing experiment and may or may not include that in the standard build.
Also did some geocoding work from OSM. Hit's ~500 r/s, but unstable so won't be included for now.

## Releases

Diglet uses a go library for reading mbtiles that depends on some C code, this makes
crosscompiling a bit of a pain, so only linux binaries will be published until there
is a need for more platforms. If a major release is it, diglet will be published for
most platforms.

We'll include a script and or directions for building on your own.

#### Minor releases will be built for linux x64 and i386

#### Major releases will have platform specific binaries

## Building

Here are the basics, but you can inspect the .travis.yml for specifics

Prereqs

* [go installed](https://golang.org/doc/install#install), preferably 1.4+
* gcc access
* go get github.com/buckhx/diglet
* cd into diglet 

```
CGO_ENABLED=1
go test -v ./...
go get ./...
go generate
go build -ldflags '-extldflags "-static"'
```

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

I got lazy and haven't maintained the licenses...
