/*
Copyright 2014 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

var ws;

jsPlumb.bind('ready', function() {
	ws = new WebSocket('ws://localhost:8080/socket');
	ws.onopen = onOpen;
	ws.onmessage = onMessage;
	ws.onclose = function() {
		console.log('socket closed');
	};
});

function send(msg) {
	ws.send(JSON.stringify(msg));
}

var plumb;

function onOpen() {
	plumb = jsPlumb.getInstance({Container: 'page'});
	plumb.bind('connection', function(conn) {
		var input = conn.targetEndpoint.getParameter('input');
		send({Action: 'connect', From: conn.source.id, To: conn.target.id, Input: input});
	});
	plumb.bind('connectionDetached', function(conn) {
		var input = conn.targetEndpoint.getParameter('input');
		send({Action: 'disconnect', From: conn.source.id, To: conn.target.id, Input: input});
	});
	plumb.bind('click', function(conn, e) {
		if (e.shiftKey) plumb.detach(conn);
	});
}

function onMessage(msg) {
	var m = JSON.parse(msg.data);

	switch (m.Action) {
		case 'hello':
			var k;
			var o = m.ObjectInputs;
			for (k in o) {
				addKind(k, o[k]);
			}
		break;
	}
}

function addKind(kind, inputs) {
	$('<li></li>').text(kind).data('inputs', inputs)
		.appendTo('#objects')
		.draggable({
			revert: true, revertDuration: 0,
			helper: 'clone',
			stop: function(e, ui) {
				newObject(kind, inputs, ui.position);
			}
		});
}

var kCount = 0;

var endpointCommon = {
	endpoint: "Dot",
	paintStyle: {
		strokeStlye: "#FFFFFF",
		fillStyle: "#FFFFFF",
		radius: 6
	},
	connector: ["Flowchart"],
	connectorStyle: {
		lineWidth: 2,
		strokeStyle: "#FFFFFF",
	}
};

function newObject(kind, inputs, offset) {
	kCount++;

	var name = kind + kCount;
	var value = 0;
	var div = $('<div class="object"></div>')
		.text(kind)
		.attr('id', name)
		.data('kind', kind)
		.appendTo('#page')
		.css('top', offset.top).css('left', offset.left)
	plumb.draggable(div);

	if (kind == "value") {
		value = window.prompt("Value? (-1 to +1)")*1;
		div.text(value).click(function(e) {
			if (!e.shiftKey) return;
			var v = window.prompt("Value? (-1 to +1)")*1;
			send({Action: 'set', Name: name, Value: v});
			$(this).text(v);
			plumb.repaint(this);
		});
	}

	// add input and output endpoints
	plumb.doWhileSuspended(function() {
		if (inputs) {
			for (var i = 0; i < inputs.length; i++) {
				plumb.addEndpoint(div, {
					uuid: name + '-' + inputs[i],
					parameters: {input: inputs[i]},
					anchor: "ContinuousTop",
					isSource: false,
					isTarget: true,
					overlays: [
						[ 'Label', {
							label: inputs[i],
							cssClass: 'label'
						} ]
					]
				}, endpointCommon);
			}
		}
		if (kind != "engine") {
			plumb.addEndpoint(div, {
				uuid: name + '-out',
				anchor: "Bottom",
				isSource: true,
				isTarget: false
			}, endpointCommon);
		}
	});

	send({Action: 'new', Name: name, Kind: kind, Value: value});
}

function demo() {
	var msgs = [
		 {Action: 'new', Name: 'engine1', Kind: 'engine'},
		 {Action: 'new', Name: 'osc1', Kind: 'osc'},
		 {Action: 'connect', From: 'osc1', To: 'engine1', Input: 'root'}
	];
	for (var i = 0; i < msgs.length; i++) {
		send(msgs[i]);
	}
	setTimeout(function() {
		send({Action: 'disconnect', From: 'osc1', To: 'engine1', Input: 'root'});
	}, 1000);
}
