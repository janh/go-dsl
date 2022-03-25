// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

"use strict";

var DSLGraphs = DSLGraphs || (function () {

	class Color {

		constructor(r, g, b, a) {
			this.r = r;
			this.g = g;
			this.b = b;
			this.a = a;
		}

		copy() {
			return new Color(this.r, this.g, this.b, this.a);
		}

		toString() {
			var a = Math.round(this.a*1000) / 1000
			return `rgba(${this.r}, ${this.g}, ${this.b}, ${a})`;
		}

	}

	Object.defineProperty(Color.prototype, 'r', {writable: true});
	Object.defineProperty(Color.prototype, 'g', {writable: true});
	Object.defineProperty(Color.prototype, 'b', {writable: true});
	Object.defineProperty(Color.prototype, 'a', {writable: true});


	class Transform {

		constructor() {
			if ("DOMMatrix" in window) {
				this._matrix = new DOMMatrix();
			} else if ("WebKitCSSMatrix" in window) {
				this._matrix = new WebKitCSSMatrix();
			} else {
				throw new Error("no suitable Matrix implementation found");
			}
		}

		scale(x, y) {
			if (y !== undefined && "scaleNonUniform" in this._matrix) {
				this._matrix = this._matrix.scaleNonUniform(x, y);
			} else {
				this._matrix = this._matrix.scale(x, y);
			}
		}

		translate(x, y) {
			this._matrix = this._matrix.translate(x, y);
		}

		abcdef() {
			return [
				this._matrix.a,
				this._matrix.b,
				this._matrix.c,
				this._matrix.d,
				this._matrix.e,
				this._matrix.f
			];
		}

	}


	class GraphParams {

		constructor() {
			this.width = 554;
			this.height = 114;
			this.scaleFactor = 1.0;
			this.fontSize = 0.0;
			this.colorBackground = new Color(255, 255, 255, 1.0);
			this.colorForeground = new Color(0, 0, 0, 1.0);
		}

	}

	Object.defineProperty(GraphParams.prototype, 'width', {writable: true});
	Object.defineProperty(GraphParams.prototype, 'height', {writable: true});
	Object.defineProperty(GraphParams.prototype, 'scaleFactor', {writable: true});
	Object.defineProperty(GraphParams.prototype, 'fontSize', {writable: true});
	Object.defineProperty(GraphParams.prototype, 'colorBackground', {writable: true});
	Object.defineProperty(GraphParams.prototype, 'colorForeground', {writable: true});


	class GraphSpec {}

	Object.defineProperty(GraphSpec.prototype, 'width', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'height', {writable: true});

	Object.defineProperty(GraphSpec.prototype, 'scaleFactor', {writable: true});

	Object.defineProperty(GraphSpec.prototype, 'colorBackground', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'colorForeground', {writable: true});

	Object.defineProperty(GraphSpec.prototype, 'legendXStep', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendXMax', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendXFactor', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendXFormatFunc', {writable: true});

	Object.defineProperty(GraphSpec.prototype, 'legendYLabelStep', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendYLabelStart', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendYLabelEnd', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendYBottom', {writable: true});
	Object.defineProperty(GraphSpec.prototype, 'legendYTop', {writable: true});


	const COLOR_UPSTREAM = Object.freeze(new Color(96, 192, 0, .75));
	const COLOR_DOWNSTREAM = Object.freeze(new Color(0, 127, 255, .75));
	const COLOR_PILOT_TONES = Object.freeze(new Color(204, 94, 82, .75));


	function decodeList(list) {
		var out = [];

		var lastVal = 0;
		var index = 0;

		while (index < list.length) {
			var cmd = list.charAt(index);
			index++;

			var numStr = "";
			while (index < list.length) {
				var c = list.charAt(index);
				if (c != '.' && (c < '0' || c > '9')) {
					break;
				}
				numStr += c;
				index++;
			}
			var num = parseInt(numStr);
			var frac = 0;
			if (numStr.length > 2 && numStr.charAt(numStr.length-2) == '.') {
				frac = parseInt(numStr.charAt(numStr.length-1));
			}

			if (cmd == 'r') {
				for (var i = 0; i < num; i++) {
					out.push(lastVal/10);
				}
				continue;
			}

			var val;

			switch (cmd) {
				case 'P': val = num * 10 + frac; break;
				case 'Q': val = num; break;
				case 'N': val = - num * 10 - frac; break;
				case 'O': val = - num; break;
				case 'p': val = lastVal + num * 10 + frac; break;
				case 'q': val = lastVal + num; break;
				case 'n': val = lastVal - num * 10 - frac; break;
				case 'o': val = lastVal - num; break;
			}

			out.push(val / 10);
			lastVal = val;
		}

		return out;
	}


	function decodeBins(data) {
		data.Bits.Downstream.Data = decodeList(data.Bits.Downstream.Data);
		data.Bits.Upstream.Data = decodeList(data.Bits.Upstream.Data);
		data.SNR.Downstream.Data = decodeList(data.SNR.Downstream.Data);
		data.SNR.Upstream.Data = decodeList(data.SNR.Upstream.Data);
		data.QLN.Downstream.Data = decodeList(data.QLN.Downstream.Data);
		data.QLN.Upstream.Data = decodeList(data.QLN.Upstream.Data);
		data.Hlog.Downstream.Data = decodeList(data.Hlog.Downstream.Data);
		data.Hlog.Upstream.Data = decodeList(data.Hlog.Upstream.Data);
		return data;
	}


	function decodeBinsHistory(data) {
		data.SNR.Downstream.Min = decodeList(data.SNR.Downstream.Min);
		data.SNR.Downstream.Max = decodeList(data.SNR.Downstream.Max);
		data.SNR.Upstream.Min = decodeList(data.SNR.Upstream.Min);
		data.SNR.Upstream.Max = decodeList(data.SNR.Upstream.Max);
		return data;
	}


	function getGraphColors(background, foreground) {
		var brightnessBackground = 0.299*background.r + 0.587*background.g + 0.114*background.b;
		var brightnessForeground = 0.299*foreground.r + 0.587*foreground.g + 0.114*foreground.b;
		var brightness = brightnessBackground;
		if (background.a < 0.75) {
			brightness = 255 - brightnessForeground;
		}

		var gray;
		if (brightness > 223) {
			gray = brightness - 20;
		} else if (brightness > 127) {
			gray = 255 - (223-brightness)/2;
		} else if (brightness > 31) {
			gray = 0 + (brightness-32)/2;
		} else {
			gray = brightness + 20;
		}

		var grayGrid;
		if (brightnessForeground < brightnessBackground) {
			grayGrid = Math.max(gray-20, 0);
		} else {
			grayGrid = Math.min(gray+20, 255);
		}

		var grayNeutral;
		if (brightness > 127) {
			grayNeutral = 95;
		} else {
			grayNeutral = 159;
		}

		return {
			colorGraph: new Color(Math.round(gray), Math.round(gray), Math.round(gray), 1.0),
			colorGrid: new Color(Math.round(grayGrid), Math.round(grayGrid), Math.round(grayGrid), 1.0),

			colorNeutralFill: new Color(Math.round(grayNeutral), Math.round(grayNeutral), Math.round(grayNeutral), .6),
			colorNeutralStroke: new Color(Math.round(grayNeutral), Math.round(grayNeutral), Math.round(grayNeutral), .75)
		}
	}


	function getLegendX(data) {
		var res = {};

		if (data) {
			res.bins = data.BinCount;
			res.freq = data.CarrierSpacing;
		} else {
			res.bins = 8192;
			res.freq = 4.3125;
		}

		switch (res.bins) {
		case 3479:
			res.step = 256;
		case 2783:
			res.step = 192;
		case 1972:
			res.step = 128;
		default:
			res.step = res.bins / 16;
		}

		return res;
	}


	class BaseGraphHelper {

		setSpec(spec) {
			var scaledWidth = spec.width / spec.scaleFactor;
			var scaledHeight = spec.height / spec.scaleFactor;

			this.width = spec.width;
			this.height = spec.height;

			var fontFactor;
			if (spec.fontSize == 0) {
				let factor = Math.min(scaledWidth/554, scaledHeight/114);
				fontFactor = Math.min(Math.max(1.0, factor), 1.35);
				this.fontSize = 10.5 * fontFactor * spec.scaleFactor;
			} else {
				fontFactor = spec.fontSize / 10.5;
				this.fontSize = spec.fontSize * spec.scaleFactor;
			}

			this.graphX = Math.round((23.0*fontFactor + 5.0) * spec.scaleFactor);
			this.graphY = Math.round(4.0 * fontFactor * spec.scaleFactor);
			this.graphWidth = spec.width - Math.round((38.0*fontFactor+4.0)*spec.scaleFactor);
			this.graphHeight = spec.height - Math.round((18.0*fontFactor+5.0)*spec.scaleFactor);

			this.colorBackground = spec.colorBackground;
			this.colorText = spec.colorForeground;

			var colors = getGraphColors(spec.colorBackground, spec.colorForeground);
			this.colorGraph = colors.colorGraph;
			this.colorGrid = colors.colorGrid;
			this.colorNeutralFill = colors.colorNeutralFill;
			this.colorNeutralStroke = colors.colorNeutralStroke;

			this.colorMinStroke = COLOR_DOWNSTREAM;
			this.colorMaxStroke = COLOR_UPSTREAM;

			this.colorUpstream = COLOR_UPSTREAM;
			this.colorDownstream = COLOR_DOWNSTREAM;

			this.colorPilotTones = COLOR_PILOT_TONES;

			if (spec.scaleFactor > 1.0) {
				this.strokeWidthBase = Math.round(spec.scaleFactor);
			} else {
				this.strokeWidthBase = 1.0;
			}

			var textOffset = 3.5 * fontFactor * spec.scaleFactor;

			var x = this.graphX;
			var y = this.graphY;
			var w = this.graphWidth;
			var h = this.graphHeight;

			var f = spec.scaleFactor;
			var ff = fontFactor;
			var s = this.strokeWidthBase;

			this._pathLegend = new Path2D();
			this._pathGrid = new Path2D();
			this._labelsX = [];
			this._labelsY = [];

			// legend for x-axis
			var legendXStep = spec.legendXStep;
			while (w*legendXStep/spec.legendXMax < this.fontSize*2.5) {
				legendXStep *= 2;
			}
			this._pathLegend.moveTo(x-0.5*s, y+h+0.5*s);
			this._pathLegend.lineTo(x-0.5*s+w, y+h+0.5*s);
			for (var i = 0.0; i <= spec.legendXMax; i += legendXStep) {
				let frac = i / spec.legendXMax;
				let pos = x - 0.5*s + Math.round(w*frac);
				this._pathLegend.moveTo(pos, y+h+Math.round(2*f)+0.5*s);
				this._pathLegend.lineTo(pos, y+h+Math.round(1*f)+0.5*s);
				let text = spec.legendXFormatFunc(i*spec.legendXFactor);
				this._labelsX.push({x: pos, y: y + h + (2+8*ff)*f + textOffset, text: text});
			}

			// legend for y-axis
			var legendYLabelStep = spec.legendYLabelStep;
			while (h*legendYLabelStep/(spec.legendYTop-spec.legendYBottom) < this.fontSize) {
				legendYLabelStep *= 2;
			}
			this.labelsYTransform = null;
			if (Math.max(Math.abs(spec.legendYLabelStart), Math.abs(spec.legendYLabelEnd)) >= 100) {
				this.labelsYTransform = new Transform();
				this.labelsYTransform.translate(x-(5+5.5*ff)*f, 0);
				this.labelsYTransform.scale(0.7, 1);
				this.labelsYTransform.translate((5+5.5*ff)*f-x, 0);
			}
			this._pathLegend.moveTo(x-0.5*s, y+0.5*s);
			this._pathLegend.lineTo(x-0.5*s, y+h+0.5*s);
			for (var i = spec.legendYLabelStart + legendYLabelStep/2; i <= spec.legendYLabelEnd; i += legendYLabelStep) {
				let frac = (i - spec.legendYBottom) / (spec.legendYTop - spec.legendYBottom);
				let pos = y + h + 0.5*s - Math.round(h*frac);
				this._pathLegend.moveTo(x-Math.round(2*f)-0.5*s, pos);
				this._pathLegend.lineTo(x-Math.round(1*f)-0.5*s, pos);
			}
			for (var i = spec.legendYLabelStart; i <= spec.legendYLabelEnd; i += legendYLabelStep) {
				let frac = (i - spec.legendYBottom) / (spec.legendYTop - spec.legendYBottom);
				let pos = y + h + 0.5*s - Math.round(h*frac);
				this._pathLegend.moveTo(x-Math.round(4*f)-0.5*s, pos);
				this._pathLegend.lineTo(x-Math.round(1*f)-0.5*s, pos);
				if (frac > 0.01) {
					this._pathGrid.moveTo(x+0.5*s, pos);
					this._pathGrid.lineTo(x+w-0.5*s, pos);
				}
				let text = i.toString();
				this._labelsY.push({x: x - (5+5.5*ff)*f, y: pos + textOffset, text: text});
			}
		}

		draw(ctx) {
			if (ctx.canvas.width != this.width || ctx.canvas.height != this.height) {
				ctx.canvas.width = this.width;
				ctx.canvas.height = this.height;
			}

			ctx.fillStyle = this.colorBackground.toString();
			ctx.fillRect(0, 0, this.width, this.height);

			ctx.fillStyle = this.colorGraph.toString();
			ctx.fillRect(this.graphX, this.graphY, this.graphWidth, this.graphHeight);

			ctx.lineWidth = this.strokeWidthBase;
			ctx.lineCap = "square";

			ctx.strokeStyle = this.colorText.toString();
			ctx.stroke(this._pathLegend);

			ctx.strokeStyle = this.colorGrid.toString();
			ctx.stroke(this._pathGrid);

			ctx.fillStyle = this.colorText.toString();
			ctx.font = this.fontSize + "px Arial,Helvetica,sans-serif";

			ctx.textAlign = "center";
			for (var item of this._labelsX) {
				ctx.fillText(item.text, item.x, item.y);
			}

			if (this.labelsYTransform != null) {
				ctx.setTransform(...this.labelsYTransform.abcdef());
			}
			ctx.textAlign = "end";
			for (var item of this._labelsY) {
				ctx.fillText(item.text, item.x, item.y);
			}
			if (this.labelsYTransform != null) {
				ctx.resetTransform();
			}
		}

	}


	class BandsGraphHelper {

		setData(data) {
			if (!data) {
				this._bands = null;
				return;
			}

			this._bins = data.BinCount;
			this._bands = [];

			for (var b of data.Bands.Downstream) {
				this._bands.push({start: b.Start, end: b.End, type: "downstream"})
			}

			for (var b of data.Bands.Upstream) {
				this._bands.push({start: b.Start, end: b.End, type: "upstream"})
			}

			this._bands.sort(function(a, b) {
				return a.start - b.start;
			});
		}

		draw(ctx, base, useColor) {
			if (this._bands == null) {
				return;
			}

			if (useColor) {
				var colorBandsDownstream = base.colorDownstream.copy();
				var colorBandsUpstream = base.colorUpstream.copy();
			} else {
				var colorBandsDownstream = base.colorNeutralFill.copy();
				var colorBandsUpstream = base.colorNeutralFill.copy();
			}
			colorBandsDownstream.a = 0.075;
			colorBandsUpstream.a = 0.075;

			var colorBandsStroke = base.colorNeutralStroke.copy();
			colorBandsStroke.a = 0.1;

			var s = base.strokeWidthBase;

			var top = base.graphY;
			var bottom = base.graphY + base.graphHeight;
			var scaleX = base.graphWidth / this._bins;

			var pathFillDownstream = new Path2D();
			var pathFillUpstream = new Path2D();
			var pathStroke = new Path2D();

			if (this._bands.length > 0) {
				var band = this._bands[0];
				var start = base.graphX + Math.floor((band.start+0.5)*scaleX);

				var pathFill = (band.type == "downstream") ? pathFillDownstream : pathFillUpstream;
				pathFill.moveTo(start, bottom);
				pathFill.lineTo(start, top);

				pathStroke.moveTo(start+0.5*s, bottom-0.5*s);
				pathStroke.lineTo(start+0.5*s, top+0.5*s);
			}

			for (var i = 1; i < this._bands.length; i++) {
				var band1 = this._bands[i-1];
				var band2 = this._bands[i];

				var end = base.graphX + Math.ceil((band1.end+0.5)*scaleX);
				var start = base.graphX + Math.floor((band2.start+0.5)*scaleX);

				if (start-end <= 1*s) {
					var center = (band2.start+band1.end) / 2;
					var pos = base.graphX + Math.floor((center+0.5)*scaleX) + 0.5*s;
					end = pos;
					start = pos;

					pathStroke.moveTo(pos, bottom-0.5*s);
					pathStroke.lineTo(pos, top+0.5*s);
				} else {
					pathStroke.moveTo(end-0.5*s, bottom-0.5*s);
					pathStroke.lineTo(end-0.5*s, top+0.5*s);

					pathStroke.moveTo(start+0.5*s, bottom-0.5*s);
					pathStroke.lineTo(start+0.5*s, top+0.5*s);
				}

				var pathFill1 = (band1.type == "downstream") ? pathFillDownstream : pathFillUpstream;
				pathFill1.lineTo(end, top);
				pathFill1.lineTo(end, bottom);
				pathFill1.closePath();

				var pathFill2 = (band2.type == "downstream") ? pathFillDownstream : pathFillUpstream;
				pathFill2.moveTo(start, bottom);
				pathFill2.lineTo(start, top);
			}

			if (this._bands.length > 0) {
				var band = this._bands[this._bands.length-1];
				var end = base.graphX + Math.ceil((band.end+0.5)*scaleX);

				var pathFill = (band.type == "downstream") ? pathFillDownstream : pathFillUpstream;
				pathFill.lineTo(end, top);
				pathFill.lineTo(end, bottom);
				pathFill.closePath();

				pathStroke.moveTo(end-0.5*s, bottom-0.5*s);
				pathStroke.lineTo(end-0.5*s, top+0.5*s);
			}

			ctx.fillStyle = colorBandsUpstream.toString();
			ctx.fill(pathFillUpstream);

			ctx.fillStyle = colorBandsDownstream.toString();
			ctx.fill(pathFillDownstream);

			ctx.lineWidth = base.strokeWidthBase;
			ctx.lineCap = "square";
			ctx.strokeStyle = colorBandsStroke.toString();
			ctx.stroke(pathStroke);
		}

	}


	function buildPilotTonesPath(path, tones, height) {
		for (var tone of tones) {
			var pos = tone + 0.5;

			path.moveTo(pos, 0);
			path.lineTo(pos, height);
		}
	}


	function buildBitsPath(path, bins, scaleY) {
		var lastValid = false;
		var lastBits = 0;
		var lastPosY = 0.0;

		var count = bins.Data.length;
		for (var i = 0; i < count; i++) {
			var bits = bins.Data[i];
			var valid = bits > 0;
			var changed = lastBits != bits;

			var posX = i;
			var posY = Math.ceil(bits * scaleY);

			if (lastValid && !valid) {
				path.lineTo(posX, lastPosY);
				path.lineTo(posX, 0);
				path.closePath();
			}
			if (!lastValid && valid) {
				path.moveTo(posX, 0);
			}
			if (valid && changed) {
				if (lastValid) {
					path.lineTo(posX, lastPosY);
				}
				path.lineTo(posX, posY);
				lastPosY = posY;
			}

			lastValid = valid;
			lastBits = bits;
		}

		if (lastValid) {
			path.lineTo(count, lastPosY);
			path.lineTo(count, 0);
			path.closePath();
		}
	}


	function buildSNRQLNPath(path, bins, scaleY, offsetY, maxY, minYValid, maxYValid) {
		var width = bins.GroupSize;

		var lastValid = false, lastDrawn = false;
		var last = offsetY;
		var lastPosY = 0.0;

		var count = bins.Data.length;
		for (var i = 0; i < count; i++) {
			var val = bins.Data[i];
			var valid = val > offsetY && val >= minYValid && val <= maxYValid;
			var changed = last != val;
			var drawn = false;

			var posX = (i + 0.5) * width;
			var posY = (Math.min(maxY, val) - offsetY) * scaleY;

			if (lastValid && !valid) {
				path.lineTo(posX-0.5*width, lastPosY);
				path.lineTo(posX-0.5*width, 0);
				path.closePath();
			}
			if (!lastValid && valid) {
				path.moveTo(posX-0.5*width, 0);
				path.lineTo(posX-0.5*width, posY);
			}
			if (valid && changed) {
				if (lastValid) {
					if (!lastDrawn) {
						path.lineTo(posX-width, lastPosY);
					}
					path.lineTo(posX, posY);
					drawn = true;
				}
				lastPosY = posY;
			}

			lastDrawn = drawn;
			lastValid = valid;
			last = val;
		}

		if (lastValid) {
			var posX = (count + 0.5) * width;
			path.lineTo(posX-0.5*width, lastPosY);
			path.lineTo(posX-0.5*width, 0);
			path.closePath();
		}
	}


	function buildSNRMinMaxPath(pathMin, pathMax, bins, scaleY, maxY, postScaleY) {
		var width = bins.GroupSize;

		var stateMin = {
			lastValid: false,
			lastDrawn: false,
			last: 0.0,
			lastPosY: 0.0
		};
		var stateMax = {
			lastValid: false,
			lastDrawn: false,
			last: 0.0,
			lastPosY: 0.0
		};

		var iter = function(path, i, val, valid, state) {
			var changed = state.last != val;
			var drawn = false;

			var posX = (i + 0.5) * width;
			var posY = Math.min(maxY, val)*scaleY - 0.5;

			if (state.lastValid && !valid) {
				path.lineTo(posX-0.5*width, state.lastPosY*postScaleY);
			}
			if (!state.lastValid && valid) {
				path.moveTo(posX-0.5*width, posY*postScaleY);
				state.lastPosY = posY;
			}
			if (valid && changed) {
				if (state.lastValid) {
					if (!state.lastDrawn) {
						path.lineTo(posX-width, state.lastPosY*postScaleY);
					}
					path.lineTo(posX, posY*postScaleY);
					drawn = true;
				}
				state.lastPosY = posY;
			}

			state.lastDrawn = drawn;
			state.lastValid = valid;
			state.last = val;
		};

		var count = bins.Min.length;
		for (var i = 0; i < count; i++) {
			var min = bins.Min[i];
			var max = bins.Max[i];
			var valid = (min > 0 && min <= 95) || (max > 0 && max <= 95);

			iter(pathMin, i, min, valid, stateMin);
			iter(pathMax, i, max, valid, stateMax);
		}

		if (stateMin.lastValid) {
			pathMin.lineTo(count*bins.GroupSize, stateMin.lastPosY*postScaleY);
		}
		if (stateMax.lastValid) {
			pathMax.lineTo(count*bins.GroupSize, stateMax.lastPosY*postScaleY);
		}
	}


	function buildHlogPath(path, bins, scaleY, offsetY, maxY, postScaleY) {
		var width = bins.GroupSize;

		var lastValid = false, lastDrawn = false;
		var last = -96.3;
		var lastPosY = 0.0;

		var count = bins.Data.length;
		for (var i = 0; i < count; i++) {
			var hlog = bins.Data[i];
			var valid = hlog >= -96.2 && hlog <= 6;
			var changed = last != hlog;
			var drawn = false;

			var posX = (i + 0.5) * width;
			var posY = Math.max(0, Math.min(maxY, hlog)-offsetY)*scaleY - 0.5;

			var reset = lastValid && Math.abs(hlog-last) >= 10;

			if ((lastValid && !valid) || reset) {
				path.lineTo(posX-0.5*width, lastPosY*postScaleY);
			}
			if ((!lastValid && valid) || reset) {
				path.moveTo(posX-0.5*width, posY*postScaleY);
				lastPosY = posY;
			}
			if (valid && changed) {
				if (lastValid && !reset) {
					if (!lastDrawn) {
						path.lineTo(posX-width, lastPosY*postScaleY);
					}
					path.lineTo(posX, posY*postScaleY);
					drawn = true;
				}
				lastPosY = posY;
			}

			lastDrawn = drawn;
			lastValid = valid;
			last = hlog;
		}

		if (lastValid) {
			path.lineTo(count*bins.GroupSize, lastPosY*postScaleY);
		}

	}


	class BitsGraph {

		constructor(canvas, params, data) {
			this._canvas = canvas;

			this._base = new BaseGraphHelper();
			this._bands = new BandsGraphHelper();

			this._spec = new GraphSpec();
			this._spec.legendXFactor = 1.0;
			this._spec.legendXFormatFunc = function(val) { return val.toFixed(0) };
			this._spec.legendYBottom = 0;
			this._spec.legendYTop = 15.166666667;
			this._spec.legendYLabelStart = 0;
			this._spec.legendYLabelEnd = 15;
			this._spec.legendYLabelStep = 2;

			this._specChanged = true;

			this._setParams(params);
			this._setData(data);

			this._draw();
		}

		_draw() {
			if (this._specChanged) {
				this._base.setSpec(this._spec);
				this._specChanged = false;
			}

			var ctx = this._canvas.getContext("2d");

			this._base.draw(ctx);

			if (!this._data) {
				return;
			}

			var x = this._base.graphX;
			var y = this._base.graphY;
			var w = this._base.graphWidth;
			var h = this._base.graphHeight;

			var scaleX = w / this._spec.legendXMax;
			var scaleY = h / this._spec.legendYTop;

			this._bands.draw(ctx, this._base, false);

			var pathPilotTones = new Path2D();
			var pathDownstream = new Path2D();
			var pathUpstream = new Path2D();

			var strokeWidthPilotTones = 1;
			if (scaleX < 1.5) {
				strokeWidthPilotTones = 1.5 / scaleX;
			}
			strokeWidthPilotTones *= this._spec.scaleFactor;

			buildPilotTonesPath(pathPilotTones, this._data.PilotTones, h);

			buildBitsPath(pathDownstream, this._data.Bits.Downstream, scaleY);
			buildBitsPath(pathUpstream, this._data.Bits.Upstream, scaleY);

			ctx.translate(x, y+h);
			ctx.scale(scaleX, -1);

			ctx.lineWidth = strokeWidthPilotTones;
			ctx.lineCap = "butt";
			ctx.strokeStyle = this._base.colorPilotTones.toString();
			ctx.stroke(pathPilotTones);

			ctx.fillStyle = this._base.colorUpstream.toString();
			ctx.fill(pathUpstream);

			ctx.fillStyle = this._base.colorDownstream.toString();
			ctx.fill(pathDownstream);

			ctx.resetTransform();
		}

		_setParams(params) {
			this._spec.width = params.width;
			this._spec.height =  params.height;
			this._spec.scaleFactor = params.scaleFactor;
			this._spec.fontSize = params.fontSize;
			this._spec.colorBackground = params.colorBackground;
			this._spec.colorForeground = params.colorForeground;

			this._specChanged = true;
		}

		setParams(params) {
			this._setParams(params);
			this._draw();
		}

		_setData(data) {
			if (this._data === undefined || !this._data != !data || (this._data && data &&
					this._data.BinCount != data.BinCount)) {

				var legendXData = getLegendX(data);
				this._spec.legendXMax = legendXData.bins;
				this._spec.legendXStep = legendXData.step;

				this._specChanged = true;
			}

			this._data = data;
			this._bands.setData(data);
		}

		setData(data) {
			this._setData(data);
			this._draw();
		}

	}


	class SNRGraph {

		constructor(canvas, params, data, history) {
			this._canvas = canvas;
			this._canvasMinMax = document.createElement("canvas");

			this._base = new BaseGraphHelper();
			this._bands = new BandsGraphHelper();

			this._spec = new GraphSpec();
			this._spec.legendXFormatFunc = function (val) { return val.toFixed(1) };
			this._spec.legendYBottom = 0;
			this._spec.legendYTop = 65;
			this._spec.legendYLabelStart = 0;
			this._spec.legendYLabelEnd = 65;
			this._spec.legendYLabelStep = 10;

			this._specChanged = true;

			this._setParams(params);
			this._setData(data, history);

			this._draw();
		}

		_draw() {
			if (this._specChanged) {
				this._base.setSpec(this._spec);
				this._specChanged = false;
			}

			var ctx = this._canvas.getContext("2d", {alpha: false});
			var ctxMinMax = this._canvasMinMax.getContext("2d");

			this._base.draw(ctx);

			if (!this._data) {
				return;
			}

			var x = this._base.graphX;
			var y = this._base.graphY;
			var w = this._base.graphWidth;
			var h = this._base.graphHeight;

			var scaleX = w / this._spec.legendXMax;
			var scaleY = h / this._spec.legendYTop;

			this._bands.draw(ctx, this._base, true);

			var path = new Path2D();
			var pathMin = new Path2D();
			var pathMax = new Path2D();

			buildSNRQLNPath(path, this._data.SNR.Downstream, scaleY, 0, this._spec.legendYTop, -32, 95);
			buildSNRQLNPath(path, this._data.SNR.Upstream, scaleY, 0, this._spec.legendYTop, -32, 95);

			if (this._history == null) {
				return;
			}

			buildSNRMinMaxPath(pathMin, pathMax, this._history.SNR.Downstream, scaleY, this._spec.legendYTop, 1/scaleX);
			buildSNRMinMaxPath(pathMin, pathMax, this._history.SNR.Upstream, scaleY, this._spec.legendYTop, 1/scaleX);

			ctx.translate(x, y+h);
			ctx.scale(scaleX, -1);

			ctx.fillStyle = this._base.colorNeutralFill.toString();
			ctx.fill(path);

			ctx.resetTransform();

			if (ctxMinMax.canvas.width != w || ctxMinMax.canvas.height != h) {
				ctxMinMax.canvas.width = w;
				ctxMinMax.canvas.height = h;
			}

			ctxMinMax.clearRect(0, 0, w, h);

			// scaling of y by scaleX in order to not distort the lines
			ctxMinMax.translate(0, h);
			ctxMinMax.scale(scaleX, -scaleX);

			ctxMinMax.lineWidth = this._spec.scaleFactor / scaleX;
			ctxMinMax.lineCap = "butt";

			ctxMinMax.globalCompositeOperation = "source-over";
			ctxMinMax.strokeStyle = this._base.colorMinStroke.toString();
			ctxMinMax.stroke(pathMin);

			ctxMinMax.globalCompositeOperation = "multiply";
			ctxMinMax.strokeStyle = this._base.colorMaxStroke.toString();
			ctxMinMax.stroke(pathMax);

			ctxMinMax.resetTransform();

			ctx.drawImage(ctxMinMax.canvas, x, y);
		}

		_setParams(params) {
			this._spec.width = params.width;
			this._spec.height =  params.height;
			this._spec.scaleFactor = params.scaleFactor;
			this._spec.fontSize = params.fontSize;
			this._spec.colorBackground = params.colorBackground;
			this._spec.colorForeground = params.colorForeground;

			this._specChanged = true;
		}

		setParams(params) {
			this._setParams(params);
			this._draw();
		}

		_setData(data, history) {
			if (this._data === undefined || !this._data != !data || (this._data && data &&
					(this._data.BinCount != data.BinCount || this._data.CarrierSpacing != data.CarrierSpacing))) {

				var legendXData = getLegendX(data);
				this._spec.legendXMax = legendXData.bins;
				this._spec.legendXStep = legendXData.step;
				this._spec.legendXFactor = legendXData.freq / 1000;

				this._specChanged = true;
			}

			this._data = data;
			this._history = history;
			this._bands.setData(data);
		}

		setData(data, history) {
			this._setData(data, history);
			this._draw();
		}

	}


	class QLNGraph {

		constructor(canvas, params, data) {
			this._canvas = canvas;

			this._base = new BaseGraphHelper();
			this._bands = new BandsGraphHelper();

			this._spec = new GraphSpec();
			this._spec.legendXFormatFunc = function (val) { return val.toFixed(1) };
			this._spec.legendYBottom = -160;
			this._spec.legendYTop = -69;
			this._spec.legendYLabelStart = -160;
			this._spec.legendYLabelEnd = -70;
			this._spec.legendYLabelStep = 20;

			this._specChanged = true;

			this._setParams(params);
			this._setData(data);

			this._draw();
		}

		_draw() {
			if (this._specChanged) {
				this._base.setSpec(this._spec);
				this._specChanged = false;
			}

			var ctx = this._canvas.getContext("2d");

			this._base.draw(ctx);

			if (!this._data) {
				return;
			}

			var x = this._base.graphX;
			var y = this._base.graphY;
			var w = this._base.graphWidth;
			var h = this._base.graphHeight;

			var scaleX = w / this._spec.legendXMax;
			var scaleY = h / (this._spec.legendYTop - this._spec.legendYBottom);

			this._bands.draw(ctx, this._base, true);

			var path = new Path2D();

			buildSNRQLNPath(path, this._data.QLN.Downstream, scaleY, this._spec.legendYBottom, this._spec.legendYTop, -150, -23);
			buildSNRQLNPath(path, this._data.QLN.Upstream, scaleY, this._spec.legendYBottom, this._spec.legendYTop, -150, -23);

			ctx.translate(x, y+h);
			ctx.scale(scaleX, -1);

			ctx.fillStyle = this._base.colorNeutralFill.toString();
			ctx.fill(path);

			ctx.resetTransform();
		}

		_setParams(params) {
			this._spec.width = params.width;
			this._spec.height =  params.height;
			this._spec.scaleFactor = params.scaleFactor;
			this._spec.fontSize = params.fontSize;
			this._spec.colorBackground = params.colorBackground;
			this._spec.colorForeground = params.colorForeground;

			this._specChanged = true;
		}

		setParams(params) {
			this._setParams(params);
			this._draw();
		}

		_setData(data) {
			if (this._data === undefined || !this._data != !data || (this._data && data &&
					(this._data.BinCount != data.BinCount || this._data.CarrierSpacing != data.CarrierSpacing))) {

				var legendXData = getLegendX(data);
				this._spec.legendXMax = legendXData.bins;
				this._spec.legendXStep = legendXData.step;
				this._spec.legendXFactor = legendXData.freq / 1000;

				this._specChanged = true;
			}

			this._data = data;
			this._bands.setData(data);
		}

		setData(data) {
			this._setData(data);
			this._draw();
		}

	}


	class HlogGraph {

		constructor(canvas, params, data) {
			this._canvas = canvas;

			this._base = new BaseGraphHelper();
			this._bands = new BandsGraphHelper();

			this._spec = new GraphSpec();
			this._spec.legendXFormatFunc = function (val) { return val.toFixed(1) };
			this._spec.legendYBottom = -100;
			this._spec.legendYTop = 7;
			this._spec.legendYLabelStart = -100;
			this._spec.legendYLabelEnd = 0;
			this._spec.legendYLabelStep = 20;

			this._specChanged = true;

			this._setParams(params);
			this._setData(data);

			this._draw();
		}

		_draw() {
			if (this._specChanged) {
				this._base.setSpec(this._spec);
				this._specChanged = false;
			}

			var ctx = this._canvas.getContext("2d");

			this._base.draw(ctx);

			if (!this._data) {
				return;
			}

			var x = this._base.graphX;
			var y = this._base.graphY;
			var w = this._base.graphWidth;
			var h = this._base.graphHeight;

			var scaleX = w / this._spec.legendXMax
			var scaleY = h / (this._spec.legendYTop - this._spec.legendYBottom)

			this._bands.draw(ctx, this._base, true);

			var path = new Path2D();

			buildHlogPath(path, this._data.Hlog.Downstream, scaleY, this._spec.legendYBottom, this._spec.legendYTop, 1/scaleX);
			buildHlogPath(path, this._data.Hlog.Upstream, scaleY, this._spec.legendYBottom, this._spec.legendYTop, 1/scaleX);

			// scaling of y by scaleX in order to not distort the line
			ctx.translate(x, y+h);
			ctx.scale(scaleX, -scaleX);

			ctx.lineWidth = this._spec.scaleFactor / scaleX;
			ctx.lineCap = "butt";
			ctx.strokeStyle = this._base.colorNeutralStroke.toString();
			ctx.stroke(path);

			ctx.resetTransform();
		}

		_setParams(params) {
			this._spec.width = params.width;
			this._spec.height =  params.height;
			this._spec.scaleFactor = params.scaleFactor;
			this._spec.fontSize = params.fontSize;
			this._spec.colorBackground = params.colorBackground;
			this._spec.colorForeground = params.colorForeground;

			this._specChanged = true;
		}

		setParams(params) {
			this._setParams(params);
			this._draw();
		}

		_setData(data) {
			if (this._data === undefined || !this._data != !data || (this._data && data &&
					(this._data.BinCount != data.BinCount || this._data.CarrierSpacing != data.CarrierSpacing))) {

				var legendXData = getLegendX(data);
				this._spec.legendXMax = legendXData.bins;
				this._spec.legendXStep = legendXData.step;
				this._spec.legendXFactor = legendXData.freq / 1000;

				this._specChanged = true;
			}

			this._data = data;
			this._bands.setData(data);
		}

		setData(data) {
			this._setData(data);
			this._draw();
		}

	}


	return {
		decodeBins: decodeBins,
		decodeBinsHistory: decodeBinsHistory,
		Color: Color,
		GraphParams: GraphParams,
		BitsGraph: BitsGraph,
		SNRGraph: SNRGraph,
		QLNGraph: QLNGraph,
		HlogGraph: HlogGraph
	}

})();
