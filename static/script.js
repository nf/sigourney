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

var ws = new WebSocket('ws://localhost:8080/socket');
ws.onopen = function() {
	demo();
	jsPlumb.bind('ready', init);
};
ws.onclose = function() {
	console.log('socket closed');
};
function send(msg) {
	ws.send(JSON.stringify(msg));
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

function init() {
	$('#objects li').draggable({
		revert: true, revertDuration: 0,
		helper: "clone",
		stop: function(e, ui) {
			newObject($(this).text().trim(), ui.position);
		}
	});
}

var kCount = 0;

function newObject(kind, offset) {
	var name = kind + kCount;
	$('<div class="object"></div>')
		.text(kind)
		.attr('name', name)
		.appendTo('#page')
		.css('top', offset.top).css('left', offset.left)
		.draggable();
	send({Action: 'new', Name: name, Kind: kind});
}
