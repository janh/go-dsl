(function() {
	"use strict";

	const STATE_READY = "ready";
	const STATE_PASSWORD = "password";
	const STATE_PASSPHRASE = "passphrase";
	const STATE_ENCRYPTION_PASSPHRASE = "encryption-passphrase";
	const STATE_ERROR = "error";
	const STATE_LOADING = "loading";
	const STATE_INITIALIZING = "initializing";
	const STATE_DISCONNECTING = "disconnecting";
	const STATE_CONNECT = "connect";

	const TRISTATE_MAYBE = 0;
	const TRISTATE_NO = -1;
	const TRISTATE_YES = 1;

	const AuthTypePassword = 1 << 0;
	const AuthTypePrivateKeys = 1 << 1;

	const OptionTypeString = 0;
	const OptionTypeBool = 1;
	const OptionTypeEnum = 2;

	var state;
	var clientDescs;

	var eventSource;

	var buttonSave, buttonDisconnect;
	var summary, graphs;
	var graphBitsCanvas, graphSNRCanvas, graphQLNCanvas, graphHlogCanvas;
	var graphBits, graphSNR, graphQLN, graphHlog;
	var overlay, overlayPassword, overlayPassphrase, overlayEncryptionPassphrase, overlayError, overlayLoading, overlayDisconnecting, overlayConnect;
	var configAdvanced, configDeviceType, configHost, configUser, configPrivateKey, configKnownHosts, configOptions, configRemember;
	var messages;
	var fingerprint, inputPassword, inputPassphrase, inputEncryptionPassphrase;

	function setConfig(config, clients) {
		clientDescs = clients;

		while (configDeviceType.firstChild) {
			configDeviceType.removeChild(configDeviceType.firstChild);
		}
		for (let key in clients) {
			let clientDesc = clients[key];
			configDeviceType.add(new Option(clientDesc.Title, key));
		}

		configDeviceType.value = config.DeviceType;
		updateConfig();

		configHost.value = config.Host;
		configUser.value = config.User;
		configPrivateKey.value = config.PrivateKeyPath;
		configKnownHosts.value = config.KnownHostsPath;

		for (let option in config.Options) {
			let id = "config-option-" + option;
			let input = document.getElementById(id);
			if (input) {
				if (input.type == "checkbox") {
					input.checked = config.Options[option] == "1";
				} else {
					input.value = config.Options[option];
				}
			}
		}
	}

	function updateConfig() {
		let deviceType = configDeviceType.value;
		let clientDesc = clientDescs[deviceType];

		if (!clientDesc) {
			clientDesc = {
				"RequiresUser": TRISTATE_MAYBE,
				"SupportedAuthTypes": 0,
				"RequiresKnownHosts": false,
				"Options": null
			};
		}

		if (clientDesc.RequiresUser == TRISTATE_NO) {
			configUser.value = "";
		}
		configUser.closest("p").classList.toggle("hide", clientDesc.RequiresUser == TRISTATE_NO);

		let hidePrivateKey = !(clientDesc.SupportedAuthTypes & AuthTypePrivateKeys);
		configPrivateKey.closest("p").classList.toggle("hide", hidePrivateKey);

		let hideKnownHosts = !clientDesc.RequiresKnownHosts;
		configKnownHosts.closest("p").classList.toggle("hide", hideKnownHosts);

		let hideOptions = !clientDesc.Options;
		configOptions.classList.toggle("hide", hideOptions);

		configAdvanced.classList.toggle("hide", hidePrivateKey && hideKnownHosts && hideOptions);

		let existingOptionItems = {};
		while (configOptions.firstChild) {
			let item = configOptions.removeChild(configOptions.firstChild);
			existingOptionItems[item.dataset.option] = item;
		}
		for (let option in clientDesc.Options) {
			let params = clientDesc.Options[option];
			let unique = JSON.stringify(params);

			let item = existingOptionItems[option];
			if (!item || item.dataset.unique != unique) {
				let id = "config-option-" + option;

				item = document.createElement("p");
				item.dataset.option = option;
				item.dataset.unique = unique;

				let label = document.createElement("label");
				item.appendChild(label);

				let title = document.createElement("span");
				title.classList.add("title");
				title.innerText = option + ":";
				label.appendChild(title);

				let input;

				switch (params.Type) {

					case OptionTypeBool:
						input = document.createElement("input");
						input.type = "checkbox";
						break;

					case OptionTypeEnum:
						input = document.createElement("select");
						for (let val of params.Values) {
							input.appendChild(new Option(val.Title, val.Value));
						}
						break;

					default:
						input = document.createElement("input");
						input.type = "text";
						input.autocomplete = "off";
						input.spellcheck = false;

				}

				input.id = id;
				input.name = option;
				label.appendChild(input);

				let desc = document.createElement("span");
				desc.classList.add("desc");
				desc.innerText = params.Description;
				label.appendChild(desc);
			}

			configOptions.appendChild(item);
		}
	}

	function connect(event) {
		var data = {
			"DeviceType": configDeviceType.value,
			"Host": configHost.value,
			"User": configUser.value,
			"PrivateKeysPath": configPrivateKey.value,
			"KnownHostsPath": configKnownHosts.value,
			"Options": {}
		};

		let optionInputs = configOptions.querySelectorAll("input,select");
		for (let input of optionInputs) {
			if (input.type == "checkbox") {
				if (input.checked) {
					data.Options[input.name] = "1";
				}
			} else {
				if (input.value.length) {
					data.Options[input.name] = input.value;
				}
			}
		}

		let remember = configRemember.checked;

		goConnect(data, remember);

		event.preventDefault();
	}

	function toggleFieldset(event) {
		let fieldset = event.target.parentElement;
		fieldset.classList.toggle("collapsed");
		event.preventDefault();
	}

	function toggleFieldsetKeyboard(event) {
		if (event.keyCode == 13 || event.keyCode == 32) {
			return toggleFieldset(event);
		}
	}

	function initConfig() {
		configDeviceType.addEventListener("change", updateConfig);
		overlayConnect.addEventListener("submit", connect);

		let legends = overlayConnect.getElementsByTagName("legend");
		for (let legend of legends) {
			legend.addEventListener("click", toggleFieldset);
			legend.addEventListener("keydown", toggleFieldsetKeyboard);
		}
	}

	function setLinkDisabled(element, disabled) {
		element.classList.toggle("disabled", disabled);
		element.tabIndex = disabled ? -1 : 0;
	}

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
					break;

			}
		}

		if (newState != oldState) {
			state = newState;

			setLinkDisabled(buttonSave, state != STATE_READY);
			setLinkDisabled(buttonDisconnect,
				state != STATE_READY && state != STATE_PASSWORD && state != STATE_PASSPHRASE && state != STATE_ENCRYPTION_PASSPHRASE && state != STATE_ERROR && state != STATE_LOADING);

			overlay.classList.toggle("visible", state != STATE_READY);
			overlayPassword.classList.toggle("visible", state == STATE_PASSWORD);
			overlayPassphrase.classList.toggle("visible", state == STATE_PASSPHRASE);
			overlayEncryptionPassphrase.classList.toggle("visible", state == STATE_ENCRYPTION_PASSPHRASE);
			overlayError.classList.toggle("visible", state == STATE_ERROR);
			overlayLoading.classList.toggle("visible", state == STATE_LOADING || state == STATE_INITIALIZING);
			overlayDisconnecting.classList.toggle("visible", state == STATE_DISCONNECTING);
			overlayConnect.classList.toggle("visible", state == STATE_CONNECT);

			if (state == STATE_PASSWORD) {
				inputPassword.focus();
			} else if (state == STATE_PASSPHRASE) {
				inputPassphrase.focus();
			} else if (state == STATE_ENCRYPTION_PASSPHRASE) {
				inputEncryptionPassphrase.focus();
			}
		}
	}

	function showMessage(text) {
		var msg = document.createElement("div");
		msg.innerText = text;

		messages.classList.add("visible");
		messages.appendChild(msg);

		window.setTimeout(function() {
			cleanupMessage(msg)
		}, 5000);
	}

	function cleanupMessage(msg) {
		msg.remove();

		if (messages.childElementCount == 0) {
			messages.classList.remove("visible");
		}
	}

	function sendForm(event) {
		let form = event.target;
		let formData = new FormData(form);
		form.reset();

		updateState("loading");

		var target = form.dataset.target;
		window[target](formData.get("data"));

		event.preventDefault();
	}

	function initForms() {
		let forms = document.querySelectorAll("#overlay-password, #overlay-passphrase, #overlay-encryption-passphrase");

		for (let form of forms) {
			form.addEventListener("submit", sendForm);
		}
	}

	function getGraphParams(width, devicePixelRatio) {
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

	function initGraphs() {
		graphBitsCanvas = document.getElementById("graph_bits");
		graphSNRCanvas = document.getElementById("graph_snr");
		graphQLNCanvas = document.getElementById("graph_qln");
		graphHlogCanvas = document.getElementById("graph_hlog");

		var defaultParams = new DSLGraphs.GraphParams();

		graphBits = new DSLGraphs.BitsGraph(graphBitsCanvas, defaultParams);
		graphSNR = new DSLGraphs.SNRGraph(graphSNRCanvas, defaultParams);
		graphQLN = new DSLGraphs.QLNGraph(graphQLNCanvas, defaultParams);
		graphHlog = new DSLGraphs.HlogGraph(graphHlogCanvas, defaultParams);

		var lastDevicePixelRatio = 0;
		var lastWidth = 0;
		var lastWidthErrors = 0;

		var updateGraphs = function() {
			var devicePixelRatio = window.devicePixelRatio;
			var width = graphs.offsetWidth;

			if (devicePixelRatio != lastDevicePixelRatio || width != lastWidth) {
				lastDevicePixelRatio = devicePixelRatio;
				lastWidth = width;

				var params = getGraphParams(width, devicePixelRatio);
				applyGraphParams(params);
			}
		};

		updateGraphs();
		window.addEventListener("resize", updateGraphs);
	}

	function initVisibilityChange() {
		var visibilityChangeListener = function() {
			goVisibilityChanged(document.visibilityState === "visible");
		};

		window.addEventListener("visibilitychange", visibilityChangeListener);

		window.addEventListener("beforeunload", function() {
			window.removeEventListener("visibilitychange", visibilityChangeListener);
		});

		visibilityChangeListener();
	}

	function loaded(event) {
		buttonSave = document.getElementById("button-save");
		buttonDisconnect = document.getElementById("button-disconnect");

		summary = document.getElementById("summary");
		graphs = document.getElementById("graphs");

		overlay = document.getElementById("overlay");
		overlayPassword = document.getElementById("overlay-password");
		overlayPassphrase = document.getElementById("overlay-passphrase");
		overlayEncryptionPassphrase = document.getElementById("overlay-encryption-passphrase");
		overlayError = document.getElementById("overlay-error");
		overlayLoading = document.getElementById("overlay-loading");
		overlayDisconnecting = document.getElementById("overlay-disconnecting");
		overlayConnect = document.getElementById("overlay-connect");

		configAdvanced = document.getElementById("config-advanced");
		configDeviceType = document.getElementById("config-device-type");
		configHost = document.getElementById("config-host");
		configUser = document.getElementById("config-user");
		configPrivateKey = document.getElementById("config-private-key");
		configKnownHosts = document.getElementById("config-known-hosts");
		configOptions = document.getElementById("config-options");
		configRemember = document.getElementById("config-remember");

		messages = document.getElementById("messages");

		fingerprint = document.getElementById("fingerprint");
		inputPassword = document.getElementById("password");
		inputPassphrase = document.getElementById("passphrase");
		inputEncryptionPassphrase = document.getElementById("encryption-passphrase");

		updateState(STATE_INITIALIZING);

		window.updateState = function(data) {
			updateState(data.state, data.data);
		}

		window.showMessage = function(text) {
			showMessage(text);
		}

		window.setConfig = function(config, clients) {
			setConfig(config, clients);
		}

		initConfig();
		initForms();
		initGraphs();
		initVisibilityChange();
		goInitialized();
	}

	document.addEventListener("DOMContentLoaded", loaded);

})();
