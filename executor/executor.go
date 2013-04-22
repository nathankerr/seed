package executor

import (
	"fmt"
	"github.com/nathankerr/seed"
	"strings"
	"time"
)

type MessageContainer struct {
	Operation  string // "immediate", "deferred", "data", "done"
	Collection string
	Data       []seed.Tuple
}

type MonitorMessage struct {
	Block string
	Data  interface{}
}

func (mc *MessageContainer) String() string {
	tuples := []string{}
	for _, tuple := range mc.Data {
		tuples = append(tuples, tuple.String())
	}
	return fmt.Sprintf("{%s %s [%s]}",
		mc.Collection, mc.Operation, strings.Join(tuples, ", "))
}

// A concurrent seed executor
// Collection and rule handlers work as concurrent processes
// managed by the control loop in this function.
func Execute(s *seed.Seed, sleepDuration time.Duration, address string, monitor bool) Channels {
	// launch the handlers
	channels := makeChannels(s)
	for collectionName, _ := range s.Collections {
		go collectionHandler(collectionName, s, channels)
	}
	for ruleNumber, _ := range s.Rules {
		go handleRule(ruleNumber, s, channels)
	}

	// make list of all processes to be controlled
	toControl := []chan<- MessageContainer{channels.Distribution}
	for _, collectionChannel := range channels.Collections {
		toControl = append(toControl, collectionChannel)
	}
	for _, ruleChannel := range channels.Rules {
		toControl = append(toControl, ruleChannel)
	}

	go controlLoop(monitor, sleepDuration, toControl, channels)
	return channels
}

func controlLoop(monitor bool, sleepDuration time.Duration, toControl []chan<- MessageContainer, channels Channels) {

	shouldStop := false

	for {
		if monitor {
			select {
			case message := <-channels.Command:
				fmt.Println(message, shouldStop)
				switch message.Data.(string) {
				case "run":
					shouldStop = false
				case "stop":
					shouldStop = true
				}
			default:
				// no-op
			}

			for shouldStop {
				channels.Monitor <- MonitorMessage {
					Block: "_command",
					Data: "stopped",
				}

				message := <-channels.Command
				fmt.Println("shouldStop", message)
				switch message.Data.(string) {
				case "run":
					shouldStop = false
				case "stop":
					// no-op
				}
			}

			channels.Monitor <- MonitorMessage {
					Block: "_command",
					Data: "running",
				}
		}

		startTime := time.Now()
		time.Sleep(sleepDuration)

		// phase 1: execute immediate rules
		messages := sendAndWaitTilFinished(
			MessageContainer{Operation: "immediate"},
			toControl, channels.Control)

		// phase 2: execute deferred rules
		messages = sendAndWaitTilFinished(
			MessageContainer{Operation: "deferred"},
			toControl, channels.Control)

		if monitor {
			for _, message := range messages {
				channels.Monitor <- MonitorMessage{
					Block: message.Collection,
					Data:  message.Data,
				}
			}

			channels.Monitor <- MonitorMessage{
				Block: "_time",
				Data:  time.Since(startTime),
			}
		}
	}
}

func sendAndWaitTilFinished(message MessageContainer, toControl []chan<- MessageContainer, controlChannel <-chan MessageContainer) []MessageContainer {
	messages := []MessageContainer{}
	sendToAll(message, toControl)
	for finished := 0; finished < len(toControl); finished++ {
		message := <-controlChannel
		messages = append(messages, message)
		controlinfo("execute", finished, "of", len(toControl))
	}
	return messages
}

func sendToAll(message MessageContainer, channels []chan<- MessageContainer) {
	for _, channel := range channels {
		channel <- message
	}
}
