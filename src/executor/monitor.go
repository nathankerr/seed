package executor

import (
	"code.google.com/p/go.net/websocket"
	// "encoding/json"
	"html/template"
	"io"
	"net/http"
	"fmt"
)

type monitorMessage struct {
	operation string
	data      interface{}
}

var registerSocket = make(chan socket)

type socket struct {
	io.ReadWriter
	done chan bool
}

var monitorAddress string

func sendToAllSockets(data []byte, sockets []socket) []int {
	toRemove := []int{}

	// send data to all sockets
	for i, socket := range sockets {
		_, err := socket.Write(data)
		if err != nil {
			// close the handler
			socket.done <- true

			// mark the (closed) socket for removal
			toRemove = append(toRemove, i)
		}
	}

	return toRemove
}

func removeSockets(toRemove []int, sockets []socket) []socket {
	// removed sockets marked for removal
	if len(toRemove) != 0 {
		// build a slice of slices for the remaining sockets
		remaining := [][]socket{}
		start := 0
		for _, end := range toRemove {
			remaining = append(remaining, sockets[start:end])
			start = end + 1
		}

		// rebuild the slice of sockets
		sockets = []socket{}
		for _, remainingSockets := range remaining {
			for _, socket := range remainingSockets {
				sockets = append(sockets, socket)
			}
		}
	}

	return sockets
}

func startMonitor(address string, channel chan monitorMessage) {
	monitorAddress = address
	go monitorServer(address)

	sockets := []socket{}
	for {
		select {
		case message := <-channel:
			monitorinfo("_monitor", message)
			data := []byte(fmt.Sprintf("MESSAGE %s %v", message.operation, message.data))
			toRemove := sendToAllSockets(data, sockets)
			sockets = removeSockets(toRemove, sockets)
		case socket := <-registerSocket:
			sockets = append(sockets, socket)
		}
	}
}

func monitorServer(address string) {
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fatal("_monitor", err)
	}
}

func socketHandler(ws *websocket.Conn) {
	done := make(chan bool)
	registerSocket <- socket{ws, done}
	<-done
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	rootTemplate.Execute(w, monitorAddress)
}

var rootTemplate = template.Must(template.New("root").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>

var output, websocket, blockArea

function showMessage(m) {
	var p = document.createElement("p")
	p.innerHTML = m
	output.appendChild(p)
	// keep the output in view
	output.scrollTop = output.scrollHeight
}

function onMessage(e) {
	showMessage(e.data)
}

function onClose() {
	showMessage("Connection Closed")
}

function newBlock(title) {
	var block = document.createElement("div")
	block.className = "block"

	var blockTitle = document.createElement("div")
	blockTitle.className = "blockTitle"
	blockTitle.innerHTML = title
	block.appendChild(blockTitle)

	var blockContent = document.createElement("div")
	blockContent.className = "blockContent"
	block.appendChild(blockContent)

	return block
}

function openBlock() {
	showMessage("creating block")
	output.style.height = (window.innerHeight / 3) + "px"

	var top = (window.innerHeight / 3)
	blockArea.style.top = top + "px"
	blockArea.style.left = "0px"
	blockArea.style.width = window.innerWidth + "px"
	blockArea.style.height = window.innerHeight - top + "px"

	var blockTitle = document.getElementById("newBlockName").value
	var block = newBlock(blockTitle)
	blockArea.appendChild(block)

	var numberOfBlocks = blockArea.children.length
	var blockWidth = window.innerWidth / numberOfBlocks - 2
	var blockHeight = window.innerHeight - top - 1

	for (i = 0; i < blockArea.children.length; i++) {
		blockArea.children[i].style.width = blockWidth - 2 +  "px"
		blockArea.children[i].style.height = blockHeight + "px"
		blockArea.children[i].style.left = ((blockWidth + 1) * i) + "px"
	}

}

function init() {
	output = document.getElementById("output")
	blockArea = document.getElementById("blockArea")

	websocket = new WebSocket("ws://{{.}}/socket");
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;

	output.style.width = window.innerWidth + "px"
	output.style.height = window.innerHeight + "px"
	output.style.top = "0px"
	output.style.left = "0px"

	showMessage("Started")
}

window.addEventListener("load", init, false);
</script>

<style>
div {
	overflow: auto;
	position: absolute;
	z-index: 0;
}

.block {
	top: 0px;
	position: absolute;
	padding: 1px;
	border: solid 1px;
}

.blockTitle {
	background-color: black;
	color: white;
	text-align: center;
	float: left;
	position: relative
	z-index: 1000;
}

.blockContent {
	padding: 0;
	margin: 0;
	position: absolute;
	left: 0;
	top: 0;
	width: 100%;
	height: 100%;
	z-index: 0;
	overflow: scroll;
}


#control {
	position: relative;
	float: right;
	z-index: 1000;
}
</style>
</head>
<body>
<div id="control">
	<select id="newBlockName">
		<option value="hello">hello</option>
		<option value="goodbye">goodbye</option>
	</select>
	<input type="button" value="Create block" onclick="openBlock()" />
</div>
<div id="output"></div>
<div id="blockArea"></div>
</body>
</html>
`))
