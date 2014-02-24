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
	$('#status').text('Connecting to back end...');
	ws = new WebSocket('ws://localhost:8080/socket');
	ws.onopen = onOpen;
	ws.onmessage = onMessage;
	ws.onclose = function() {
		$('#status').text('Lost connection to back end!');
	};
});

function send(msg) {
	ws.send(JSON.stringify(msg));
}

var plumb;

function onOpen() {
	$('#status').empty();
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

	var fn = $('<input type="text"/>');
	var save = $('<input type="button" value="save"/>');
	var load = $('<input type="button" value="load"/>');
	$('#control').append(fn, load, save);

	load.click(function() {
		$('.object').each(function() { plumb.remove(this); });
		send({Action: 'load', Name: fn.val()});
	});
	save.click(function() {
		send({Action: 'save', Name: fn.val()});
	});
}

function onMessage(msg) {
	var m = JSON.parse(msg.data);

	switch (m.Action) {
		case 'hello':
			handleHello(m.KindInputs);
			break;
		case 'new':
			bumpNCount(m.Name);
			newObjectName(m.Name, m.Kind, m.Value, m.Display);
			break;
		case 'connect':
			plumb.connect({uuids: [m.From + '-out', m.To + '-' + m.Input]});
			break;
	}
}

var kindInputs = {};

function handleHello(inputs) {
	var k;
	for (k in inputs) {
		kindInputs[k] = inputs[k];
		if (k == "engine") {
			newObject(k, {offset: engineOffset()});
		} else {
			addKind(k, inputs[k]);
		}
	}
}

function engineOffset() {
	var w = $('#page').width(), h = $('#page').height();
	return {top: 3*h/4, left: w/2-(49+20+20+1+1)/2};
}

function addKind(kind, inputs) {
	$('<li></li>').text(kind).data('inputs', inputs)
		.appendTo('#objects')
		.draggable({
			revert: true, revertDuration: 0,
			helper: 'clone',
			stop: function(e, ui) {
				newObject(kind, {offset: ui.position});
			}
		});
}

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

var nCount = 0;

function bumpNCount(name) {
	var r = /[0-9]+$/.exec(name);
	if (r === null) return; // unknown name
	var n = r[0]*1;
	if (nCount <= n) {
		nCount = n + 1;
	}
}

function newObject(kind, display) {
	nCount++;
	var name = kind + nCount;
	if (kind == "engine")
		name = "engine";
	newObjectName(name, kind, null, display)
}

function newObjectName(name, kind, value, display) {
	var div = $('<div class="object"></div>')
		.text(kind)
		.attr('id', name)
		.data('kind', kind)
		.appendTo('#page')
		.css('top', display.offset.top)
		.css('left', display.offset.left)
	var setDisplay = function() {
		send({Action: 'setDisplay', Name: name, Display: {offset: $(div).offset()}});
	}
	plumb.draggable(div, { stop: setDisplay });

	if (kind == "value") {
		if (value === null) value = 0;
		div.text(value).dblclick(function(e) {
			var v = window.prompt("Value? (-1 to +1)")*1;
			send({Action: 'set', Name: name, Value: v});
			$(this).text(v);
			plumb.repaint(this);
		});
	}

	if (kind != "engine") {
		div.click(function(e) {
			if (!e.shiftKey) return;
			plumb.remove(this);
			send({Action: 'destroy', Name: name});
		});
	}

	// add input and output endpoints
	plumb.doWhileSuspended(function() {
		var inputs = kindInputs[kind];
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
				isTarget: false,
				maxConnections: -1
			}, endpointCommon);
		}
	});

	if (kind != "engine")
		send({Action: 'new', Name: name, Kind: kind, Value: value});
	setDisplay();
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
