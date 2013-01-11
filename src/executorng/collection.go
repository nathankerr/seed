package executorng

import (
	"service"
)

func collectionHandler(collectionName string, s *service.Service, channels channels) {
	controlinfo(collectionName, "started")
	input := channels.collections[collectionName]

	for {
		message := <-input
		controlinfo(collectionName, "received", message)
		channels.finished <- true
		controlinfo(collectionName, "finished with", message)
	}
}
