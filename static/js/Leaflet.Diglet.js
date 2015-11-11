L.TileLayer.DigletSource = L.TileLayer.extend({

	initialize: function (url, tileset) {
		//L.TileLayer.prototype.initialize.call(this, "", {});
		this._tileset = tileset; //TODO assert url and tileset
		this._wsBind(url);
	},

	_wsBind: function(url) {
		that = this;
		ws = new WebSocket(url); //TODO assert ws/wss
		ws.onopen = function(e) {
			while (that._wsOpenqueue.length > 0) {
			    that._wsSend(that._wsOpenqueue.pop());
			}
			that._wsOnOpen(e);
		};
		ws.onmessage = function(e) { 
			msg = JSON.parse(e.data)
			that._wsOnMessage(msg); 
		}; 
		ws.onclose = this._wsOnClose
		ws.onerror = this._wsOnError
		this._wsOpenqueue = [];
		this._wsUrl = url;
		this._ws = ws;
		this._wsTiles = {};
		return ws;
	},

	_wsRpc: function(method, id, params) {
		id = id || Math.floor(Math.random() * (4294967295 - 0)); // (0, max_uint]
		params = params || {};
		this._wsSend({id: id, method: method, params: params, jsonrpc: "2.0"});
	},

	_wsSend: function(msg) {
		if (!this._wsIsOpen()) {
			this._wsOpenqueue.push(msg);
		} else {
			this._ws.send(JSON.stringify(msg))
		}
	},

	_wsOnMessage: function(e) {
		if ('error' in e) {
			console.log(e);
		} else if ('id' in e) { // is RPC
			if (e.id in this._wsTiles) {
				tile = this._wsTiles[e.id]
				tile.src = 'data:image/png;base64,' + e.result.data
			};
		} else {
			console.log(e);
		}
	},

	_wsOnOpen: function(e) {},
	
	_wsOnClose: function(e) {
		delete this._ws;
		delete this._wsTiles;
	},
	
	_wsOnError: function(e) {
		console.log(e);
	},

	_wsIsOpen: function() {
		return this._ws && this._ws.readyState === 1;
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
			tileset: this._tileset,
		}
		key = [params.z, params.x, params.y].join(":")
		this._wsTiles[key] = tile;
		this._wsRpc('get_tile', key, params);
		//this._wsRpc('subscribe_tile', key, params);

		this.fire('tileloadstart', {
			tile: tile,
			url: tile.src
		});
		return tile;
	},
});
