package executorng

import (
	"service"
)

func budCommunicator(s *service.Service, channels channels) {
	controlinfo("budCommunicator", "started")
	for {
		message := <-channels.distribution
		controlinfo("budCommunicator", "received", message)

		switch message.operation {
		case "immediate", "deferred":
			channels.finished <- true
			controlinfo("budCommunicator", "finished with", message)
		default:
			fatal("budCommunicator", "unhandled message:", message)
		}
	}
}
