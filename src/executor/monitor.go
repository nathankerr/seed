package executor

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
)

type monitorMessage struct {
	Block string
	Data  interface{}
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
			// data := []byte(fmt.Sprintf("MESSAGE %s %v", message.operation, message.data))
			data, err := json.Marshal(message)
			if err != nil {
				panic(err)
			}
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

var websocket, focus, blocks, knownBlockNames

function showMessage(m) {
	var p = document.createElement("p")
	p.innerHTML = m

	var logBlock = document.getElementById("log")
	if (logBlock == null) {
		return
	}

	var log = logBlock.children[1]
	if (log == null) {
		return
	}

	log.appendChild(p)
	// keep the output in view
	log.scrollTop = log.scrollHeight
}

function onMessage(e) {
	var message = JSON.parse(e.data)
	
	knownBlockNames[message.Block] = true
	setNewBlockNames()

	var block = document.getElementById(message.Block)
	if (block != null) {
		block.children[1].innerHTML = message.Data
	}

	// showMessage(message)
}

function setNewBlockNames() {
	var newBlockName = document.getElementById("newBlockName")
	var names = Object.keys(knownBlockNames)
	if (newBlockName.children.length != names.length) {
		for (var i = newBlockName.children.length; i > 0; i--) {
			var block = newBlockName.children[0]
			newBlockName.removeChild(block)
		}

		for (i=0; i < names.length; i++) {
			var name = names[i]

			var option = document.createElement("option")
			option.value = name
			option.innerHTML = name

			newBlockName.appendChild(option)
		}

	}
}

function onClose() {
	showMessage("Connection Closed")
}

function newBlock(title) {
	var block = document.createElement("div")
	block.className = "block"
	block.id = title

	var blockTitle = document.createElement("div")
	blockTitle.className = "blockTitle"
	blockTitle.innerHTML = title
	blockTitle.onclick = focusBlock
	block.appendChild(blockTitle)

	var blockContent = document.createElement("div")
	blockContent.className = "blockContent"
	block.appendChild(blockContent)

	var blockClose = document.createElement("div")
	blockClose.className = "blockClose"
	blockClose.innerHTML = "[x]"
	blockClose.onclick = closeBlock
	block.appendChild(blockClose)

	return block
}

function createBlock(blockTitle) {
	showMessage("creating block")

	// var blockTitle = document.getElementById("newBlockName").value
	var block = newBlock(blockTitle)
	var content = block.children[1]

	if (focus.children.length == 0) {
		// first block
		focus.appendChild(block)
	} else {
		blocks.appendChild(block)
	}
	resizeBlocks()

	content.style.height = Number(block.style.height.match("[0-9]+")[0]) - 40 + "px"
}

function resizeBlocks() {
	// focused
	var focused = focus.children[0]
	if (focused != null) {
		focused.style.height = focus.style.height;
		focused.style.top = "0px"
	}

	// blocks
	var numberOfBlocks = blocks.children.length
	var blockHeight = window.innerHeight / numberOfBlocks

	for (i = 0; i < blocks.children.length; i++) {
		blocks.children[i].style.height = blockHeight + "px"
		blocks.children[i].style.top = ((blockHeight + 1) * i) + "px"
	}
}

function closeBlock() {
	var block = this.parentElement
	var container = block.parentElement
	
	container.removeChild(block)

	resizeBlocks()
}

function focusBlock(block) {
	var block = this.parentElement
	var container = block.parentElement

	if (container.id == "focus") {
		return
	}

	var focused = focus.children[0]
	if (focused != null) {
		focus.removeChild(focused)
		blocks.appendChild(focused)
	}

	focus.appendChild(block)

	resizeBlocks()
}

function init() {
	knownBlockNames = {}
	knownBlockNames["log"] = true

	websocket = new WebSocket("ws://{{.}}/socket");
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;

	focus = document.getElementById("focus")
	var focusWidth = window.innerWidth * 0.618
	focus.style.width = focusWidth + "px"
	focus.style.height = window.innerHeight + "px"

	control = document.getElementById("control")
	control.style.left = focusWidth + "px"
	control.style.width = window.innerWidth - focusWidth + "px"

	blocks = document.getElementById("blocks")
	blocks.style.left = focusWidth + "px"
	blocks.style.width = window.innerWidth - focusWidth + "px"
	blocks.style.height = window.innerHeight + "px"

	createBlock("log")

	showMessage("started")
}

window.addEventListener("load", init, false);
</script>

<style>
div {
	overflow: auto;
	position: absolute;
	z-index: 0;
}

#focus {
	top: 0px;
	left: 0px;
}

#blocks {
	top: 20px;
}

.block {
	left: 0px;
	overflow: hidden;
	width: 100%;
}

.blockTitle {
	background-color: black;
	color: white;
	text-align: center;
	top: 0px;
	left: 0px;
	width: 100%;
	height: 20px;
}

.blockClose {
	left: 0px;
	z-index: 1;
	width: 30px;
	height: 20px;
	background-color: black;
	color: white;
}

.blockContent {
	left: 0px;
	top: 20px;
	width: 100%;
	overflow: scroll;
}

#control {
	top: 0px;
}

</style>
</head>
<body>
<div id="control">
	<select id="newBlockName"></select>
	<input type="button" value="Open" onclick="createBlock(document.getElementById('newBlockName').value)" />
</div>
<div id="focus"></div>
<div id="blocks">
</div>
</body>
</html>
`))
