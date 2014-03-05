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

var Sigourney = {};

Sigourney.UI = function() {
	var ui = this;

	var ws;
	var plumb;

	var changedSinceSave = false;
	var changeWarning = "There are unsaved changes!\nOK to continue?";

	jsPlumb.bind('ready', function() {
		$('#status').text('Connecting to back end...');
		ws = new WebSocket('ws://localhost:8080/socket');
		ws.onopen = onOpen;
		ws.onmessage = onMessage;
		ws.onclose = function() {
			$('#status').append('<div>Lost connection to back end!</div>');
		};
	});

	function send(m) {
		console.log(">", m);
		ws.send(JSON.stringify(m));
	}

	function onOpen() {
		$('#status').empty();

		initPlumb();
		initUI();
	}

	function initPlumb() {
		plumb = jsPlumb.getInstance({Container: 'page'});
		plumb.bind('connection', function(conn) {
			var input = conn.targetEndpoint.getParameter('input');
			send({Action: 'connect', From: conn.source.id, To: conn.target.id, Input: input});
			changedSinceSave = true;
		});
		plumb.bind('connectionDetached', function(conn) {
			if (!conn.targetEndpoint.isTarget) return;
			var input = conn.targetEndpoint.getParameter('input');
			send({Action: 'disconnect', From: conn.source.id, To: conn.target.id, Input: input});
			changedSinceSave = true;
		});
		plumb.bind('click', function(conn, e) {
			if (!e.shiftKey) return;
			plumb.detach(conn);
		});
	}

	function setValue(obj, v) {
		$(obj).data('value', v).text(v);
		plumb.repaint($(obj));
		send({Action: 'set', Name: $(obj).attr('id'), Value: v});
		changedSinceSave = true;
	}

	function destroy(obj) {
		plumb.remove($(obj));
		send({Action: 'destroy', Name: $(obj).attr('id')});
		changedSinceSave = true;
	}

	function initUI() {
		var fn = $('<input type="text"/>');
		var load = $('<input type="button" value="load"/>');
		var save = $('<input type="button" value="save"/>');
		$('#control').append(fn, load, save);

		var loadFn  = function() {
			if (changedSinceSave && !confirm(changeWarning)) return;
			$('.object').each(function() { plumb.remove(this); });
			send({Action: 'load', Name: fn.val()});
			fn.blur();
			load.blur();
			changedSinceSave = false;
		};
		fn.keypress(function(e) { if (e.charCode == 13) loadFn(); });
		load.click(loadFn);
		save.click(function() {
			send({Action: 'save', Name: fn.val()});
			save.blur();
			changedSinceSave = false;
		});

		// Handle keypresses while fn field is not focused.
		$(document).keypress(function(e) {
			if (fn.is(':focus'))
				return;
			switch (e.charCode) {
			case 100: // d
				onDup();
				break;
			case 24: // ^x
				onDelete();
				break;
			default:
				return;
			}
			e.preventDefault();
		})

		// Blur inputs on clicks outside controls.
		$('#page, #objects').mousedown(function(e) {
			if (!$(e.originalEvent.target).is('#control input')) {
				$('#control input').blur();
				console.log('false');
			}
		});

		$('#page').selectable({filter: ".object"})
	}

	function onMessage(msg) {
		var m = JSON.parse(msg.data);
		console.log("<", m);
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
			case 'message':
				var div = $('<div></div>').text(m.Message);
				$('#status').append(div);
				setTimeout(function() { div.remove(); }, 5000);
				break;
		}
	}

	var kindInputs = {};
	var colorIndex = 0;

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
			.addClass('kind-'+kind)
			.appendTo('#objects')
			.draggable({
				revert: true, revertDuration: 0,
				helper: 'clone',
				stop: function(e, ui) {
					newObject(kind, {offset: ui.position});
					changedSinceSave = true;
				}
			});
	}

	var endpointCommon = {
		endpoint: "Dot",
		paintStyle: {
			fillStyle: "#DDD",
			radius: 6
		},
		hoverPaintStyle: {
			 fillStyle: "#FFF"
		},
		connector: ["Flowchart"],
		connectorStyle: {
			lineWidth: 3,
			strokeStyle: "#BBB"
		},
		connectorHoverStyle: {
			strokeStyle: "#FFF"
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
		return name;
	}

	function sendDisplay(obj) {
		send({
			Action: 'setDisplay',
			Name: $(obj).attr('id'),
			Display: {offset: $(obj).offset()}
		});
	}

	function newObjectName(name, kind, value, display) {
		var div = $('<div class="object"></div>')
			.text(kind)
			.attr('id', name)
			.data('kind', kind)
			.css('top', display.offset.top)
			.css('left', display.offset.left)
			.addClass('kind-'+kind)
			.appendTo('#page')

		plumb.draggable(div, {
			start: function(e, ui) {
				if (!$(this).is('.ui-selected')) return;
			},
			drag: function(e, ui) {
				if (!$(this).is('.ui-selected')) return;
				var o1 = $(this).offset();
				var p = ui.position;
				$('.ui-selected').not(this).each(function() {
					var o2 = $(this).offset();
					$(this).css({
						top: p.top-o1.top+o2.top,
						left: p.left-o1.left+o2.left
					});
					plumb.repaint(this);
				});
			},
			stop: function(e, ui) {
				sendDisplay(this);
				if (!$(this).is('.ui-selected')) return;
				$('.ui-selected').not(this).each(function() {
					plumb.repaint(this);
				}).each(function() {
					sendDisplay(this);
				});
			}
		});

		if (kind == "value") {
			if (value === null) value = 0;
			div.data('value', value).text(value).dblclick(function(e) {
				setValue(this, window.prompt("Value? (-1 to +1)")*1);
			});
		}

		if (kind != "engine") {
			div.click(function(e) {
				if (!e.shiftKey) return;
				destroy(this);
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
		sendDisplay(div);
	}

	function onDup() {
		var names = {};
		var added = {};
		$('.ui-selected').removeClass('ui-selected').not('#engine').each(function() {
			changedSinceSave = true;
			names[$(this).attr('id')] = true;
		}).each(function() {
			var o = $(this).offset();
			o.top += 50;
			o.left += 50;
			var k = $(this).data('kind')
			var n = newObject(k, {offset: o});
			names[$(this).attr('id')] = n;
			var el = $('#'+n).addClass('ui-selected');
			if (k == 'value') {
				setValue(el, $(this).data('value'));
			}
		}).each(function() {
			var id = $(this).attr('id');
			var conns = plumb.getConnections($(this));
			for (var i = 0; i < conns.length; i++) {
				var c = conns[i];
				if (c.sourceId != id)
					continue;
				var source = names[c.sourceId];
				var target = names[c.targetId];
				if (!target)
					continue;
				var input;
				for (var j = 0; j < c.endpoints.length; j++) {
					var e = c.endpoints[j];
					if (e.elementId == c.targetId)
						input = e.getParameter('input');
				}
				plumb.connect({uuids: [source + '-out', target + '-' + input]});
			}
		});
	}

	function onDelete() {
		$('.ui-selected').not('#engine').each(function() { destroy(this); });
	}
};

var ui = new Sigourney.UI;
