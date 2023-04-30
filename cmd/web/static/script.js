"use strict";

const STATE_READY = "ready";
const STATE_PASSWORD = "password";
const STATE_PASSPHRASE = "passphrase";
const STATE_ENCRYPTION_PASSPHRASE = "encryption-passphrase";
const STATE_ERROR = "error";
const STATE_LOADING = "loading";

var state;

var eventSource;

var summary, graphs, errors;
var checkboxAutoscale;
var graphBitsCanvas, graphSNRCanvas, graphQLNCanvas, graphHlogCanvas,
	graphRetransmissionDownCanvas, graphRetransmissionUpCanvas,
	graphErrorsDownCanvas, graphErrorsUpCanvas,
	graphErrorSecondsDownCanvas, graphErrorSecondsUpCanvas;
var graphBits, graphSNR, graphQLN, graphHlog,
	graphRetransmissionDown, graphRetransmissionUp,
	graphErrorsDown, graphErrorsUp,
	graphErrorSecondsDown, graphErrorSecondsUp;
var overlay, overlayPassword, overlayPassphrase, overlayEncryptionPassphrase, overlayError, overlayLoading;
var fingerprint, inputPassword, inputPassphrase, inputEncryptionPassphrase;

function updateState(newState, data) {
	let oldState = state;

	if (data !== undefined) {
		switch (newState) {

			case STATE_PASSPHRASE:
				fingerprint.innerText = data;
				break;

			case STATE_READY:
				var bins = DSLGraphs.decodeBins(data["bins"]);
				var binsHistory = DSLGraphs.decodeBinsHistory(data["bins_history"]);
				var errorsHistory = DSLGraphs.decodeErrorsHistory(data["errors_history"]);
				summary.innerHTML = data["summary"];
				graphBits.setData(bins);
				graphSNR.setData(bins, binsHistory);
				graphQLN.setData(bins);
				graphHlog.setData(bins);
				graphRetransmissionDown.setData(errorsHistory);
				graphRetransmissionUp.setData(errorsHistory);
				graphErrorsDown.setData(errorsHistory);
				graphErrorsUp.setData(errorsHistory);
				graphErrorSecondsDown.setData(errorsHistory);
				graphErrorSecondsUp.setData(errorsHistory);
				break;

			case STATE_ERROR:
				overlayError.innerText = data;
				break;

		}
	}

	if (newState != oldState) {
		state = newState;

		checkboxAutoscale.disabled = state != STATE_READY;

		overlay.classList.toggle("visible", state != STATE_READY);
		overlayPassword.classList.toggle("visible", state == STATE_PASSWORD);
		overlayPassphrase.classList.toggle("visible", state == STATE_PASSPHRASE);
		overlayEncryptionPassphrase.classList.toggle("visible", state == STATE_ENCRYPTION_PASSPHRASE);
		overlayError.classList.toggle("visible", state == STATE_ERROR);
		overlayLoading.classList.toggle("visible", state == STATE_LOADING);

		if (state == STATE_PASSWORD) {
			inputPassword.focus();
		} else if (state == STATE_PASSPHRASE) {
			inputPassphrase.focus();
		} else if (state == STATE_ENCRYPTION_PASSPHRASE) {
			inputEncryptionPassphrase.focus();
		}
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

function getGraphParams(width, devicePixelRatio, autoscale) {
	var params = new DSLGraphs.GraphParams();

	var width = width;
	if (width >= params.width) {
		var height = params.height + 0.75 * (width-params.width) * params.height / params.width;
	} else {
		var height = params.height;
	}

	params.width = Math.floor(width * devicePixelRatio);
	params.height = Math.floor(height * devicePixelRatio);

	params.scaleFactor = devicePixelRatio;

	params.preferDynamicAxisLimits = autoscale;

	return params;
}

function applyGraphParams(params) {
	var width = (params.width / params.scaleFactor).toString() + "px";

	graphBits.setParams(params);
	graphBitsCanvas.style.width = width;

	graphSNR.setParams(params);
	graphSNRCanvas.style.width = width;

	graphQLN.setParams(params);
	graphQLNCanvas.style.width = width;

	graphHlog.setParams(params);
	graphHlogCanvas.style.width = width;
}

function applyErrorsGraphParams(params) {
	var width = (params.width / params.scaleFactor).toString() + "px";

	graphRetransmissionDown.setParams(params);
	graphRetransmissionDownCanvas.style.width = width;

	graphRetransmissionUp.setParams(params);
	graphRetransmissionUpCanvas.style.width = width;

	graphErrorsDown.setParams(params);
	graphErrorsDownCanvas.style.width = width;

	graphErrorsUp.setParams(params);
	graphErrorsUpCanvas.style.width = width;

	graphErrorSecondsDown.setParams(params);
	graphErrorSecondsDownCanvas.style.width = width;

	graphErrorSecondsUp.setParams(params);
	graphErrorSecondsUpCanvas.style.width = width;
}

function initGraphs() {
	graphBitsCanvas = document.getElementById("graph_bits");
	graphSNRCanvas = document.getElementById("graph_snr");
	graphQLNCanvas = document.getElementById("graph_qln");
	graphHlogCanvas = document.getElementById("graph_hlog");

	graphRetransmissionDownCanvas = document.getElementById("graph_retransmission_ds");
	graphRetransmissionUpCanvas = document.getElementById("graph_retransmission_us");
	graphErrorsDownCanvas = document.getElementById("graph_errors_ds");
	graphErrorsUpCanvas = document.getElementById("graph_errors_us");
	graphErrorSecondsDownCanvas = document.getElementById("graph_errorseconds_ds");
	graphErrorSecondsUpCanvas = document.getElementById("graph_errorseconds_us");

	var defaultParams = new DSLGraphs.GraphParams();

	graphBits = new DSLGraphs.BitsGraph(graphBitsCanvas, defaultParams);
	graphSNR = new DSLGraphs.SNRGraph(graphSNRCanvas, defaultParams);
	graphQLN = new DSLGraphs.QLNGraph(graphQLNCanvas, defaultParams);
	graphHlog = new DSLGraphs.HlogGraph(graphHlogCanvas, defaultParams);

	graphRetransmissionDown = new DSLGraphs.DownstreamRetransmissionGraph(graphRetransmissionDownCanvas, defaultParams);
	graphRetransmissionUp = new DSLGraphs.UpstreamRetransmissionGraph(graphRetransmissionUpCanvas, defaultParams);
	graphErrorsDown = new DSLGraphs.DownstreamErrorsGraph(graphErrorsDownCanvas, defaultParams);
	graphErrorsUp = new DSLGraphs.UpstreamErrorsGraph(graphErrorsUpCanvas, defaultParams);
	graphErrorSecondsDown = new DSLGraphs.DownstreamErrorSecondsGraph(graphErrorSecondsDownCanvas, defaultParams);
	graphErrorSecondsUp = new DSLGraphs.UpstreamErrorSecondsGraph(graphErrorSecondsUpCanvas, defaultParams);

	var lastDevicePixelRatio = 0;
	var lastWidth = 0;
	var lastWidthErrors = 0;
	var lastAutoscale = false;

	var updateGraphs = function() {
		var devicePixelRatio = window.devicePixelRatio;

		var autoscale = checkboxAutoscale.checked;
		var width = graphs.offsetWidth;
		if (devicePixelRatio != lastDevicePixelRatio || width != lastWidth || autoscale != lastAutoscale) {
			lastAutoscale = autoscale;
			lastWidth = width;

			var params = getGraphParams(width, devicePixelRatio, autoscale);
			applyGraphParams(params);
		}

		var widthErrors = errors.offsetWidth;
		if (widthErrors > 1000) {
			widthErrors = Math.floor(widthErrors/2) - 1;
		}
		if (devicePixelRatio != lastDevicePixelRatio || widthErrors != lastWidthErrors) {
			lastWidthErrors = widthErrors;

			var paramsErrors = getGraphParams(widthErrors, devicePixelRatio, false);
			applyErrorsGraphParams(paramsErrors);
		}

		lastDevicePixelRatio = devicePixelRatio;
	};

	updateGraphs();
	window.addEventListener("resize", updateGraphs);
	checkboxAutoscale.addEventListener("change", updateGraphs);
}

function loaded(event) {
	summary = document.getElementById("summary");
	graphs = document.getElementById("graphs");
	errors = document.getElementById("errors");

	checkboxAutoscale = document.getElementById("checkbox-autoscale");

	overlay = document.getElementById("overlay");
	overlayPassword = document.getElementById("overlay-password");
	overlayPassphrase = document.getElementById("overlay-passphrase");
	overlayEncryptionPassphrase = document.getElementById("overlay-encryption-passphrase");
	overlayError = document.getElementById("overlay-error");
	overlayLoading = document.getElementById("overlay-loading");

	fingerprint = document.getElementById("fingerprint");
	inputPassword = document.getElementById("password");
	inputPassphrase = document.getElementById("passphrase");
	inputEncryptionPassphrase = document.getElementById("encryption-passphrase");

	updateState("loading");

	initForms();
	initGraphs();
	initEvents();
}

document.addEventListener("DOMContentLoaded", loaded);
