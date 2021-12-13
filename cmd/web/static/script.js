"use strict";

const STATE_READY = "ready";
const STATE_PASSWORD = "password";
const STATE_PASSPHRASE = "passphrase";
const STATE_ERROR = "error";
const STATE_LOADING = "loading";

var state;

var eventSource;

var summary, graphBits, graphSNR, graphQLN, graphHlog;
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
				summary.innerText = data["summary"];
				graphBits.src = "data:image/svg+xml," + encodeURIComponent(data["graph_bits"]);
				graphSNR.src = "data:image/svg+xml," + encodeURIComponent(data["graph_snr"]);
				graphQLN.src = "data:image/svg+xml," + encodeURIComponent(data["graph_qln"]);
				graphHlog.src = "data:image/svg+xml," + encodeURIComponent(data["graph_hlog"]);
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

function loaded(event) {
	summary = document.getElementById("summary");
	graphBits = document.getElementById("graph_bits");
	graphSNR = document.getElementById("graph_snr");
	graphQLN = document.getElementById("graph_qln");
	graphHlog = document.getElementById("graph_hlog");

	overlay = document.getElementById("overlay");
	overlayPassword = document.getElementById("overlay-password");
	overlayPassphrase = document.getElementById("overlay-passphrase");
	overlayError = document.getElementById("overlay-error");
	overlayLoading = document.getElementById("overlay-loading");

	fingerprint = document.getElementById("fingerprint");

	updateState("loading");

	initForms();
	initEvents();
}

document.addEventListener("DOMContentLoaded", loaded);
