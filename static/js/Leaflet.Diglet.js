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
		req = {id: id, method: method, params: params, jsonrpc: "2.0"};
		this.send(req);
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

var Protobuf = Pbf;

L.TileLayer.DigletMVTSource = L.TileLayer.MVTSource.extend({ 

	initialize: function (url, tileset, options) {
		layer = this;
		options = options || {};
		L.TileLayer.MVTSource.prototype.initialize.call(layer, options);
		layer._wsTiles = {};
		layer._wsTileset = tileset;
		layer._wsRpc = new RpcWebSocket(url, {
			message: function(e) {
				if ('error' in e) {
					console.log(e);
				} else if ('id' in e) {
					if (e.id in layer._wsTiles) {
						//console.log(layer._wsTiles);
						ctx = layer._wsTiles[e.id]
						ctx.canvas.setAttribute("id", e.id)
						// We're not using the tile.data here
						// It gets loaded from _draw with an xhr
						layer.drawTile(ctx.canvas, ctx.tile, ctx.zoom);
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
		layer.on('tileunload', function(e) {
			var id = e.tile.id;
			params = {
				z: Number(id.split(":")[0]),
				x: Number(id.split(":")[1]),
				y: Number(id.split(":")[2]),
				tileset: layer._wsTileset,
			}
			layer._wsRpc.request('unsubscribe_tile', "unsub:"+id, params);
			delete layer._wsTiles[e.tile.id]
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
		var ctx = {
			id: [params.z, params.x, params.y].join(":"),
			canvas: tile,
			tile: coords,
			zoom: params.z,
			tileSize: this.options.tileSize
		};
		this._wsTiles[ctx.id] = ctx;
		ctx.canvas.setAttribute("id", ctx.id)
		//this._wsRpc.request('get_tile', ctx.id, params);
		// Draw what's currently available
		layer.drawTile(ctx.canvas, ctx.tile, ctx.zoom);
		this._wsRpc.request('subscribe_tile', "sub:"+ctx.id, params);

		this.fire('tileloadstart', {
			tile: tile,
			url: tile.src
		});
		return tile;
	},

	/*
	_onTileRemove: function (e) {
		console.log(e);
		e.tile.onload = null;
		//this._wsRpc.request('subscribe_tile', "sub:"+ctx.id, params);
	},
	*/
});
