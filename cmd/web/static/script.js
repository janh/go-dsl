"use strict";

const STATE_READY = "ready";
const STATE_PASSWORD = "password";
const STATE_PASSPHRASE = "passphrase";
const STATE_ERROR = "error";
const STATE_LOADING = "loading";

var state;

var eventSource;

var summary, graphs;
var graphBitsCanvas, graphSNRCanvas, graphQLNCanvas, graphHlogCanvas;
var graphBits, graphSNR, graphQLN, graphHlog;
var overlay, overlayPassword, overlayPassphrase, overlayError, overlayLoading;
var fingerprint;

function updateState(newState, data) {
	let oldState = state;

	if (data !== undefined) {
		switch (newState) {

			case STATE_PASSPHRASE:
				fingerprint.innerText = data;
				break;

			case STATE_READY:
				var bins = DSLGraphs.decodeBins(data["bins"]);
				var history = DSLGraphs.decodeBinsHistory(data["history"]);
				summary.innerHTML = data["summary"];
				graphBits.setData(bins);
				graphSNR.setData(bins, history);
				graphQLN.setData(bins);
				graphHlog.setData(bins);
				break;

			case STATE_ERROR:
				overlayError.innerText = data;

		}
	}

	if (newState != oldState) {
		state = newState;

		overlay.classList.toggle("visible", state != STATE_READY);
		overlayPassword.classList.toggle("visible", state == STATE_PASSWORD);
		overlayPassphrase.classList.toggle("visible", state == STATE_PASSPHRASE);
		overlayError.classList.toggle("visible", state == STATE_ERROR);
		overlayLoading.classList.toggle("visible", state == STATE_LOADING);
	}
}

function sendForm(event) {
	let form = event.target;
	let formData = new FormData(form);
	form.reset();

	updateState("loading");

	fetch(form.action, {
		method: form.method,
		body: formData
	})
	.then(response => {
		if (!response.ok) {
			updateState("error", "failed to send input");
		}
	})
	.catch(error => {
		updateState("error", "failed to send input");
	});

	event.preventDefault();
}

function initForms() {
	let forms = document.getElementsByTagName("form");

	for (let form of forms) {
		form.addEventListener("submit", sendForm);
	}
}

function startEvents() {
	eventSource = new EventSource("events");

	eventSource.onmessage = function(event) {
		let message = JSON.parse(event.data);
		updateState(message.state, message.data);
	}

	eventSource.onerror = function(event) {
		updateState("error", "disconnected from server");
	}
}

function stopEvents() {
	eventSource.close();
}

function initEvents() {
	startEvents();

	document.addEventListener("visibilitychange", function() {
		if (document.visibilityState === "visible") {
			updateState("loading");
			startEvents();
		} else {
			stopEvents();
		}
	});
}

function getGraphParams(width, devicePixelRatio) {
	var params = new DSLGraphs.GraphParams();

	var width = width;
	var height = 114;

	params.width = Math.round(width * devicePixelRatio);
	params.height = Math.round(height * devicePixelRatio);

	params.scaleFactor = devicePixelRatio;

	return params;
}

function setCanvasWidths(params) {
	var width = (params.width / params.scaleFactor).toString() + "px";
	graphBitsCanvas.style.width = width;
	graphSNRCanvas.style.width = width;
	graphQLNCanvas.style.width = width;
	graphHlogCanvas.style.width = width;
}

function initGraphs() {
	var lastDevicePixelRatio = window.devicePixelRatio;
	var lastWidth = graphs.offsetWidth;

	var params = getGraphParams(lastWidth, lastDevicePixelRatio);

	graphBitsCanvas = document.getElementById("graph_bits");
	graphSNRCanvas = document.getElementById("graph_snr");
	graphQLNCanvas = document.getElementById("graph_qln");
	graphHlogCanvas = document.getElementById("graph_hlog");

	graphBits = new DSLGraphs.BitsGraph(graphBitsCanvas, params);
	graphSNR = new DSLGraphs.SNRGraph(graphSNRCanvas, params);
	graphQLN = new DSLGraphs.QLNGraph(graphQLNCanvas, params);
	graphHlog = new DSLGraphs.HlogGraph(graphHlogCanvas, params);

	setCanvasWidths(params);

	window.addEventListener("resize", function() {
		var devicePixelRatio = window.devicePixelRatio;
		var width = graphs.offsetWidth;

		if (devicePixelRatio != lastDevicePixelRatio || width != lastWidth) {
			lastDevicePixelRatio = devicePixelRatio;
			lastWidth = width;

			var params = getGraphParams(width, devicePixelRatio);

			graphBits.setParams(params);
			graphSNR.setParams(params);
			graphQLN.setParams(params);
			graphHlog.setParams(params);

			setCanvasWidths(params);
		}
	});
}

function loaded(event) {
	summary = document.getElementById("summary");
	graphs = document.getElementById("graphs");

	overlay = document.getElementById("overlay");
	overlayPassword = document.getElementById("overlay-password");
	overlayPassphrase = document.getElementById("overlay-passphrase");
	overlayError = document.getElementById("overlay-error");
	overlayLoading = document.getElementById("overlay-loading");

	fingerprint = document.getElementById("fingerprint");

	updateState("loading");

	initForms();
	initGraphs();
	initEvents();
}

document.addEventListener("DOMContentLoaded", loaded);
