package executor

import (
	"service"
)

type channels struct {
	control      chan messageContainer
	distribution chan messageContainer
	collections  map[string]chan messageContainer
	rules        []chan messageContainer
}

func makeChannels(s *service.Service) channels {
	var channels channels

	channels.control = make(chan messageContainer)
	channels.distribution = make(chan messageContainer)

	channels.collections = make(map[string]chan messageContainer)
	for collectionName, _ := range s.Collections {
		channels.collections[collectionName] =
			make(chan messageContainer)
	}

	channels.rules = []chan messageContainer{}
	for _, _ = range s.Rules {
		channels.rules = append(channels.rules,
			make(chan messageContainer))
	}

	return channels
}
