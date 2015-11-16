function noop() {}

var RpcWebSocket = (function () { 

	function RpcWebSocket(url, handlers) {
		//TODO assert ws/wss
		this.url = url;
		this._openqueue = [];
		this._ws = new WebSocket(url);
		setHandlers.call(this, handlers);
	}

	RpcWebSocket.prototype.request = function(method, id, params) {
		params = params || {};
		id = id || Math.floor(Math.random() * (4294967295 - 0)); // (0, max_uint]
		this.send({id: id, method: method, params: params, jsonrpc: "2.0"});
	}

	RpcWebSocket.prototype.send = function(msg) {
		if (!this.isOpen()) {
			this._openqueue.push(msg);
		} else {
			this._ws.send(JSON.stringify(msg))
		}
	}

	RpcWebSocket.prototype.isOpen = function() {
		return this._ws && this._ws.readyState === 1;
	}

	function setHandlers(handlers) {
		that = this;
		handlers = handlers || {};
		open = handlers.open || noop;
		handlers.open = function(e) {
			while (that._openqueue.length > 0) {
				that.send(that._openqueue.pop());
			}
			open(e);
		}
		message = handlers.message || noop;
		handlers.message = function(e) {
			msg = JSON.parse(e.data);
			message(msg);
		};
		handlers.close = handlers.close || noop;
		handlers.error = handlers.error || function(e) {console.log(e);};
		this._ws.onopen = handlers.open;
		this._ws.onmessage = handlers.message;
		this._ws.onerror = handlers.error;
		this._ws.onclose = handlers.close;
		return handlers
	}

	return RpcWebSocket;
})();


L.TileLayer.DigletSource = L.TileLayer.extend({

	initialize: function (url, tileset, handlers) {
		layer = this;
		layer._wsTiles = {};
		layer._wsTileset = tileset; //TODO assert url and tileset
		layer._wsRpc = new RpcWebSocket(url, {
			message: function(e) {
				if ('error' in e) {
					console.log(e);
				} else if ('id' in e) {
					if (e.id in layer._wsTiles) {
						// TODO only set if not undefined
						tile = layer._wsTiles[e.id]
						tile.src = 'data:image/png;base64,' + e.result.data
					};
				} else {
					console.log(e);
				}
			},
			close:   function(e) {
				delete layer._ws;
				delete layer._wsTiles;
			},
		});
	},

	_loadTile: function (tile, coords) {
		tile._layer  = this;
		tile.onload  = this._tileOnLoad;
		tile.onerror = this._tileOnError;

		this._adjustTilePoint(coords);

		params = {
			r: this.options.detectRetina && L.Browser.retina && this.options.maxZoom > 0 ? '@2x' : '',
			x: coords.x,
			y: this.options.tms ? this._globalTileRange.max.y - coords.y : coords.y,
			z: this._getZoomForUrl(),
			tileset: this._wsTileset,
		}
		key = [params.z, params.x, params.y].join(":")
		this._wsTiles[key] = tile;
		this._wsRpc.request('get_tile', key, params);
		this._wsRpc.request('subscribe_tile', "sub:"+key, params);

		this.fire('tileloadstart', {
			tile: tile,
			url: tile.src
		});
		return tile;
	},
});

var VectorTile = require('vector-tile').VectorTile;
var Protobuf = require('pbf');

L.TileLayer.DigletMVTSource = L.TileLayer.MVTSource.extend({ 

	initialize: function (url, tileset, handlers, options) {
		options = options || {};
		options.url = "{x}/{y}/{z}";
		L.TileLayer.MVTSource.prototype.initialize.call(this, options);
		layer = this;
		layer._wsTiles = {};
		layer._wsTileset = tileset;
		layer._wsRpc = new RpcWebSocket(url, {
			message: function(e) {
				if ('error' in e) {
					console.log(e);
				} else if ('id' in e) {
					if (e.id in layer._wsTiles) {
						self = layer;
						var tile = e.result;
						var ctx = self._wsTiles[e.id];
						var arrayBuffer = new Uint8Array(tile.data);
						var buf = new Protobuf(arrayBuffer);
						var vt = new VectorTile(buf);
						if (self.map && self.map.getZoom() != ctx.zoom) {
							console.log("Fetched tile for zoom level " + ctx.zoom + ". Map is at zoom level " + self._map.getZoom());
						        return;
						}
						self.checkVectorTileLayers(parseVT(vt), ctx);
						tileLoaded(self, ctx);
						self.reduceTilesToProcessCount();
					}
				} else {
					console.log(e);
				}
			},
			close:   function(e) {
				delete layer._ws;
				delete layer._wsTiles;
			},
		});
	},

	_draw: function(ctx) { 
		params = {
			r: this.options.detectRetina && L.Browser.retina && this.options.maxZoom > 0 ? '@2x' : '',
			x: ctx.tile.x,
			y: this.options.tms ? this._globalTileRange.max.y - ctx.tile.y : ctx.tile.y,
			z: this._getZoomForUrl(),
			tileset: this._wsTileset,
		}
		this._wsTiles[ctx.id] = ctx;
		this._wsRpc.request('get_tile', ctx.id, params);
		this._wsRpc.request('subscribe_tile', "sub:"+ctx.id, params);
	},

});
