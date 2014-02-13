var ws = new WebSocket("ws://localhost:8080/socket");
ws.onopen = function() {
	demo();
};
ws.onclose = function() {
	console.log("socket closed");
};
function send(msg) {
	ws.send(JSON.stringify(msg));
}

jsPlumb.bind("ready", function() {

});

function demo() {
	var msgs = [
		 {Action: "new", Name: "engine1", Kind: "engine"},
		 {Action: "new", Name: "osc1", Kind: "osc"},
		 {Action: "connect", From: "osc1", To: "engine1", Input: "root"}
	];
	for (var i = 0; i < msgs.length; i++) {
		send(msgs[i]);
	}
	setTimeout(function() {
		send({Action: "disconnect", From: "osc1", To: "engine1", Input: "root"});
	}, 1000);
}
