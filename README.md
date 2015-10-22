# diglet

A tile server in a single binary.

There are some unique features in the pipeline for dynamically serving tiles, but
for now diglet is just a small server for serving mbtiles in vector or raster form.

# Usage
 
    diglet start --port 80 --mbtiles ~/path/to/dir/with/mbtiles_files/

--mbtiles: Path to local directory containing mbtiles files. NOTE only serves files
with .mbtiles extension

--port default is 8080

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

