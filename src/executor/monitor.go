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

func startMonitor(address string, channel chan monitorMessage) {
	monitorAddress = address
	go monitorServer(address)

	sockets := []socket{}
	for {
		select {
		case message := <-channel:
			toRemove := []int{}

			// send data to all sockets
			for i, socket := range sockets {
				// data, err := json.Marshal(message.data)
				// if err != nil {
				// 	panic(err)
				// }

				data := []byte(fmt.Sprintf("MESSAGE %s %v", message.operation, message.data))

				_, err := socket.Write(data)
				if err != nil {
					// close the handler
					socket.done <- true

					// mark the (closed) socket for removal
					toRemove = append(toRemove, i)
				}
			}

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

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>

var output, websocket

function showMessage(m) {
	var p = document.createElement("p")
	p.innerHTML = m
	output.appendChild(p)
}

function onMessage(e) {
	showMessage(e.data)
}

function onClose() {
	showMessage("Connection Closed")
}

function init() {
	output = document.getElementById("output")

	websocket = new WebSocket("ws://{{.}}/socket");
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;

	showMessage("Started")
}

window.addEventListener("load", init, false);
</script>
</head>
<body>
<input type="button" onclick="showMessage('clicked')" />
<div id="viz"></div>
<div id="block1"></div>
<div id="block2"></div>
<div id="block3"></div>
<div id="block4"></div>
<div id="block5"></div>
<div id="block6"></div>
</body>
</html>
`))
