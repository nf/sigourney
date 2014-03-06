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

Sigourney.Debug = false;

Sigourney.UI = function() {
	var ui = this;
	var plumb;

	ui.objects = {};
	ui.changedSinceSave = false;

	var kindInputs = {};
	var colorIndex = 0;

	jsPlumb.bind('ready', function() {
		$('#status').text('Connecting to back end...');
		ui.ws = new WebSocket('ws://localhost:8080/socket');
		ui.ws.onopen = function() {
			$('#status').empty();

			initPlumb();
			initUI();
		};
		ui.ws.onclose = function() {
			$('#status').append('<div>Lost connection to back end!</div>');
		};
		ui.ws.onmessage = onMessage;
	});

	function onMessage(msg) {
		var m = JSON.parse(msg.data);
		if (Sigourney.Debug) console.log("<", m);
		switch (m.Action) {
			case 'hello':
				handleHello(m.KindInputs);
				break;
			case 'setGraph':
				plumb.doWhileSuspended(function() {
					handleSetGraph(m.Graph);
				});
				break;
			case 'message':
				var div = $('<div></div>').text(m.Message);
				$('#status').append(div);
				setTimeout(function() { div.remove(); }, 5000);
				break;
		}
	}

	function initPlumb() {
		ui.plumb = plumb = jsPlumb.getInstance({Container: 'page'});
		plumb.bind('connection', function(conn) {
			var source = ui.objects[conn.sourceId];
			var target = ui.objects[conn.targetId];
			var input = conn.targetEndpoint.getParameter('input');
			if (target.inputs[input] != source.name) {
				target.inputs[input] = source.name;
				ui.send({Action: 'connect', From: source.name, To: target.name, Input: input});
				ui.changedSinceSave = true;
			}
		});
		plumb.bind('connectionDetached', function(conn) {
			if (!conn.targetEndpoint.isTarget) return;
			var target = ui.objects[conn.targetId];
			var source = ui.objects[conn.sourceId];
			var input = conn.targetEndpoint.getParameter('input');
			target.inputs[input] = null;
			ui.send({Action: 'disconnect', From: conn.source.id, To: conn.target.id, Input: input});
			ui.changedSinceSave = true;
		});
		plumb.bind('click', function(conn, e) {
			if (!e.shiftKey) return;
			plumb.detach(conn);
		});
	}

	function initUI() {
		var fn = $('<input type="text"/>');
		var load = $('<input type="button" value="load"/>');
		var save = $('<input type="button" value="save"/>');
		$('#control').append(fn, load, save);

		var loadFn  = function() {
			var changeWarning = "There are unsaved changes!\nOK to continue?";
			if (ui.changedSinceSave && !confirm(changeWarning)) return;
			for (var name in ui.objects) {
				ui.objects[name].destroy();
			}
			ui.objects = {};
			ui.send({Action: 'load', Name: fn.val()});
			fn.blur();
			load.blur();
			ui.changedSinceSave = false;
		};
		fn.keypress(function(e) { if (e.charCode == 13) loadFn(); });
		load.click(loadFn);
		save.click(function() {
			ui.send({Action: 'save', Name: fn.val()});
			save.blur();
			ui.changedSinceSave = false;
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
			}
		});

		$('#page').selectable({filter: ".object"})
	}

	function handleHello(inputs) {
		for (var k in inputs) {
			kindInputs[k] = inputs[k];
			if (k == "engine") {
				createObject(k, {offset: engineOffset()});
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
					createObject(kind, {offset: ui.position});
					ui.changedSinceSave = true;
				}
			});
	}

	function handleSetGraph(graph) {
		for (var i = 0; i < graph.length; i++) {
			var o = graph[i];
			bumpNCount(o.Name);
			newObject(o);
		}
		for (var i = 0; i < graph.length; i++) {
			var o = graph[i];
			for (var input in o.Input) {
				var from = o.Input[input];
				if (!from) continue;
				plumb.connect({uuids: [from + '-out', o.Name + '-' + input]});
				ui.objects[o.Name].inputs[input] = from
			}
		}
	}

	var nCount = 0;

	function bumpNCount(name) {
		var r = /[0-9]+$/.exec(name);
		if (r === null) return; // unknown name
		var n = r[0]*1;
		if (nCount <= n) {
			nCount = n + 1;
		}
	}

	function createObject(kind, display) {
		nCount++;
		var name = kind + nCount;
		if (kind == "engine")
			name = "engine";

		var obj = newObject({Name: name, Kind: kind, Display: display});

		if (kind != "engine") {
			var m = {Action: 'new', Name: name, Kind: kind};
			if (kind == "value")
				m.Value = obj.value;
			ui.send(m);
		}
		ui.onDisplayUpdate(obj);

		return obj;
	}

	function newObject(b) {
		var inputs = {};
		var kInputs = kindInputs[b.Kind];
		if (kInputs != null) {
			for (var i = 0; i < kInputs.length; i++) {
				inputs[kInputs[i]] = null;
			}
		}

		var obj = new Sigourney.Object(ui, b, inputs);
		ui.objects[b.Name] = obj;
		obj.element();

		return obj;
	}

	function onDup() {
		var names = {};
		$('.ui-selected').not('#engine').each(function() {
			// duplicate objects
			var obj1 = $(this).data('object');
			var o1 = obj1.display.offset;
			var o2 = {top: o1.top + 50, left: o1.left + 50};
			var obj2 = createObject(obj1.kind, {offset: o2});
			names[obj1.name] = obj2.name;
			obj2.element().addClass('ui-selected');
			if (obj1.kind == 'value') {
				obj2.setValue(obj1.value);
				ui.onSetValue(obj2);
			}
		}).each(function() {
			// connect new objects
			var obj = $(this).data('object');
			for (var input in obj.inputs) {
				var targetName = names[obj.name];
				var sourceName = names[obj.inputs[input]];
				if (!sourceName)
					continue;
				plumb.connect({uuids: [sourceName + '-out', targetName + '-' + input]});
			}
		}).removeClass('ui-selected');
	}

	function onDelete() {
		$('.ui-selected').not('#engine').each(function() {
			var obj = $(this).data('object');
			obj.destroy();
			ui.onDestroy(obj);
		});
	}
};

Sigourney.UI.prototype.send = function(m) {
	if (Sigourney.Debug) console.log(">", m);
	this.ws.send(JSON.stringify(m));
}

Sigourney.UI.prototype.onDisplayUpdate = function(obj, nosend) {
	this.send({Action: 'setDisplay', Name: obj.name, Display: obj.display});
	this.changedSinceSave = true;
};

Sigourney.UI.prototype.onSetValue = function(obj) {
	this.send({Action: 'set', Name: obj.name, Value: obj.value});
	this.changedSinceSave = true;
};

Sigourney.UI.prototype.onDestroy = function(obj) {
	this.send({Action: 'destroy', Name: obj.name});
	this.changedSinceSave = true;
	delete(objects[obj.name]);
};

Sigourney.Object = function(ui, b, inputs) {
	this.ui = ui;
	this.el = null;

	this.name = b.Name;
	this.kind = b.Kind;
	this.value = b.Value || 0;
	this.display = b.Display;

	this.inputs = inputs;
};

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

Sigourney.Object.prototype.element = function() {
	if (this.el != null) {
		return this.el;
	}

	var obj = this;
	var ui = obj.ui;
	var plumb = ui.plumb;

	obj.el = $('<div class="object"></div>')
		.data('object', obj)
		.attr('id', obj.name)
		.text(obj.kind).addClass('kind-'+obj.kind)
		.css('top', obj.display.offset.top)
		.css('left', obj.display.offset.left)
		.appendTo('#page')

	plumb.draggable(obj.el, {
		start: function() {
			if (!$(this).is('.ui-selected')) return;
		},
		drag: function(e, jqUI) {
			if (!$(this).is('.ui-selected')) return;
			var o1 = $(this).offset();
			var p = jqUI.position;
			$('.ui-selected').not(this).each(function() {
				var o2 = $(this).offset();
				$(this).css({
					top: p.top-o1.top+o2.top,
					left: p.left-o1.left+o2.left
				});
				plumb.repaint(this);
			});
		},
		stop: function() {
			obj.setDisplay();
			ui.onDisplayUpdate(obj);
			if (!$(this).is('.ui-selected'))
				return;
			$('.ui-selected').not(this).each(function() {
				plumb.repaint(this);
				var obj = $(this).data('object');
				obj.setDisplay();
				ui.onDisplayUpdate(obj);
			});
		}
	});

	if (obj.kind == "value") {
		obj.setValue(obj.value);
		obj.el.dblclick(function(e) {
			var v = window.prompt("Value? (-1 to +1)")*1;
			obj.setValue(v);
			ui.onSetValue(obj);
		});
	}

	if (obj.kind != "engine") {
		obj.el.click(function(e) {
			if (!e.shiftKey) return;
			obj.destroy();
			ui.onDestroy(obj);
		});
	}

	// add input and output endpoints
	plumb.doWhileSuspended(function() {
		if (obj.inputs) {
			for (var input in obj.inputs) {
				plumb.addEndpoint(obj.el, {
					uuid: obj.name + '-' + input,
					parameters: {input: input},
					anchor: "ContinuousTop",
					isSource: false,
					isTarget: true,
					overlays: [
						[ 'Label', {
							label: input,
							cssClass: 'label'
						} ]
					]
				}, endpointCommon);
			}
		}
		if (obj.kind != "engine") {
			plumb.addEndpoint(obj.el, {
				uuid: obj.name + '-out',
				anchor: "Bottom",
				isSource: true,
				isTarget: false,
				maxConnections: -1
			}, endpointCommon);
		}
	});

	return this.el;
}

Sigourney.Object.prototype.setDisplay = function() {
	this.display = {offset: this.element().offset()};
}

Sigourney.Object.prototype.setValue = function(v) {
	this.value = v;
	this.ui.plumb.repaint(this.element().text(v));
}

Sigourney.Object.prototype.destroy = function() {
	if (this.el == null) return;
	this.ui.plumb.remove($(this.el));

}

var ui = new Sigourney.UI;
