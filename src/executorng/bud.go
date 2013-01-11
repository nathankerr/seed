package executorng

import (
	"service"
)

func budCommunicator(s *service.Service, channels channels) {
	controlinfo("budCommunicator", "started")
	for {
		message := <-channels.distribution
		controlinfo("budCommunicator", "received", message)
		channels.finished <- true
		controlinfo("budCommunicator", "finished with", message)
	}
}
