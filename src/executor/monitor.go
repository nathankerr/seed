package executor

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"service"
	"strconv"
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

func sendStartupData(s *service.Service, socket socket) {
	messages := []monitorMessage{}

	// _service block content
	messages = append(messages, monitorMessage{
		Block: "_service",
		Data:  fmt.Sprintf("<code>%s</code>", s.String()[1:]), // skip the beginning newline in the string
	})

	// list of collections for input control
	collections := ""
	for name, _ := range s.Collections {
		collections += fmt.Sprintf("<option value=\"%s\">%s</option>", name, name)
	}
	messages = append(messages, monitorMessage{
		Block: "_collections",
		Data:  collections,
	})

	for _, message := range messages {
		data, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}

		socket.Write(data)
	}
}

func startMonitor(address string, channel chan monitorMessage, s *service.Service) {
	monitorAddress = address
	go monitorServer(address)

	sockets := []socket{}
	for {
		select {
		case message := <-channel:
			monitorinfo("_monitor", message)
			message.Data = renderHTML(message, s)
			data, err := json.Marshal(message)
			if err != nil {
				panic(err)
			}
			toRemove := sendToAllSockets(data, sockets)
			sockets = removeSockets(toRemove, sockets)
		case socket := <-registerSocket:
			sockets = append(sockets, socket)
			sendStartupData(s, socket)
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

func renderHTML(message monitorMessage, s *service.Service) string {
	collection, ok := s.Collections[message.Block]
	if !ok {
		number, err := strconv.ParseInt(message.Block, 0, 0)
		if err == nil {
			rule := s.Rules[number]
			collection = s.Collections[rule.Supplies]
		} else {
			switch message.Block {
			case "_time", "budCommunicator":
				return fmt.Sprint(message.Data)
			default:
				panic("unhandled block: " + message.Block)
			}
		}
	}

	table := "<table><tr>"

	// add headers
	for _, column := range collection.Key {
		table += fmt.Sprintf("<th>%s</th>", column)
	}
	for _, column := range collection.Data {
		table += fmt.Sprintf("<th>%s</th>", column)
	}
	table += "</tr>"

	rows := message.Data.([]tuple)
	for _, row := range rows {
		table += "<tr>"
		for _, column := range row {
			switch typed := column.(type) {
			case []byte:
				table += fmt.Sprintf("<td>%s</td>", string(typed))
			default:
				table += fmt.Sprintf("<td>%v</td>", column)
			}
		}
		table += "</tr>"
	}

	table += "</table>"

	return table
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	rootTemplate.Execute(w, monitorAddress)
}

var rootTemplate = template.Must(template.New("root").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>

var websocket, focus, blocks, knownBlockNames, connected

function showMessage(m) {
	var p = document.createElement("p")
	p.innerHTML = m

	var logBlock = document.getElementById("_log")
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
	
	knownBlockNames[message.Block] = message.Data
	setNewBlockNames()

	if (message.Block == "_collections") {
		document.getElementById("sendToCollection").innerHTML = message.Data
	} else {
		var block = document.getElementById(message.Block)
		if (block != null) {
			block.children[1].innerHTML = message.Data
		}
	}
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
	connected.style.backgroundColor = "red"
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
	var block = document.getElementById(blockTitle)
	if (block != null) {
		showMessage("block already open")
		return
	}

	showMessage("creating block")

	var block = newBlock(blockTitle)
	var content = block.children[1]

	if (focus.children.length == 0) {
		// first block
		focus.appendChild(block)
	} else {
		blocks.appendChild(block)
	}
	resizeBlocks()

	// add data if we have it
	var data = knownBlockNames[blockTitle]
	if (data != null) {
		content.innerHTML = data
	}
}

function resizeBlocks() {
	// focused
	var focused = focus.children[0]
	if (focused != null) {
		focused.style.height = focus.style.height;
		focused.style.top = "0px"
	}
	resizeContent(focused)

	// blocks
	var numberOfBlocks = blocks.children.length
	var blockHeight = window.innerHeight / numberOfBlocks

	for (i = 0; i < blocks.children.length; i++) {
		var block = blocks.children[i]
		block.style.height = blockHeight + "px"
		block.style.top = ((blockHeight + 1) * i) + "px"

		// set height of the content of the block
		resizeContent(block)
	}
}

function resizeContent(block) {
	var content = block.children[1]
		content.style.height = Number(block.style.height.match("[0-9]+")[0]) - 20 + "px"
}

function resizeContainers() {
	var focusWidth = window.innerWidth * 0.618
	var focusHeight = window.innerHeight * 0.618
	focus.style.width = focusWidth + "px"
	focus.style.height = focusHeight + "px"

	connected.style.top = focusHeight + "px"

	control.style.top = focusHeight + "px"
	control.style.width = focusWidth + "px"
	control.style.height = window.innerHeight - focusHeight + "px"

	blocks.style.left = focusWidth + "px"
	blocks.style.width = window.innerWidth - focusWidth + "px"
	blocks.style.height = window.innerHeight + "px"
}

function resizeAll() {
	resizeContainers()
	resizeBlocks()
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
	knownBlockNames["_log"] = ""

	// fill in the globals for the frequently accessed objs
	connected = document.getElementById("connected")
	focus = document.getElementById("focus")
	control = document.getElementById("control")
	blocks = document.getElementById("blocks")

	resizeContainers()

	// connect to the monitor server
	websocket = new WebSocket("ws://{{.}}/socket");
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;
	connected.style.backgroundColor = "green"

	// open _log
	createBlock("_log")
	showMessage("started")
}

window.addEventListener("load", init, false);
window.onresize=resizeAll
</script>

<style>
body {
	overflow: hidden;
}
div {
	overflow: hidden;
	position: absolute;
	z-index: 0;
}

#focus {
	top: 0px;
	left: 0px;
}

#blocks {
	top: 0px;
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
	overflow: auto;
}

#control {
	left: 0px;
}

#connected {
	left: 0px;
	width: 20px;
	height: 20px;
	background-color: red;
}

table {
	width: 100%;
	border-collapse: collapse;
}

td {
	text-align: center;
}

code {
	white-space: pre;
}

</style>
</head>
<body>
<div id="control" class="block">
	<div class="blockTitle">control</div>
	<div class="blockContent">
		<select id="newBlockName"></select>
		<input type="button" value="Open" onclick="createBlock(document.getElementById('newBlockName').value)" />

		<div style="position: relative;">
			<textarea id="toSend" style="position: relative; float:left; height: 36px;"></textarea>
			<div style="position: relative; float: left;">
				<select id="sendToCollection"></select>
				<br/>
				<input type="button" value="Insert" />
			</div>
		</div>

		<span id="timestep_status">Running</span>
		<input type="button" value="Immediate"/>
		<input type="button" value="Deferred"/>
		<input type="button" value="Run"/>
		<input type="button" value="Stop"/>
	</div>
</div>
<div id="connected" class="connected-red">&nbsp;</div>
<div id="focus"></div>
<div id="blocks">
</div>
</body>
</html>
`))
