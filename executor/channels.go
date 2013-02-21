package executor

import (
	service "github.com/nathankerr/seed"
)

type Channels struct {
	Control      chan MessageContainer
	Distribution chan MessageContainer
	Collections  map[string]chan MessageContainer
	Rules        []chan MessageContainer
	Monitor      chan MonitorMessage
}

func makeChannels(s *service.Seed) Channels {
	var channels Channels

	channels.Control = make(chan MessageContainer)
	channels.Distribution = make(chan MessageContainer)

	channels.Collections = make(map[string]chan MessageContainer)
	for collectionName, _ := range s.Collections {
		channels.Collections[collectionName] =
			make(chan MessageContainer)
	}

	channels.Rules = []chan MessageContainer{}
	for _, _ = range s.Rules {
		channels.Rules = append(channels.Rules,
			make(chan MessageContainer))
	}

	channels.Monitor = make(chan MonitorMessage)

	return channels
}
