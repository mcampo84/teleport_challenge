package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

// define constants for valid command inputs - start, stop, status, stream
const (
	commandStart  string = "start"
	commandStop   string = "stop"
	commandStatus string = "status"
	commandStream string = "stream"
)

type Command interface {
	Execute(context.Context) error
}

type BaseCommand struct {
	client *client.Client
}

func NewCommand(client *client.Client, input []string) (_ Command, err error) {
	// validate the input against valid commands - start, stop, status, stream.
	// then return the appropriate command
	if len(input) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	// log the directive (1st argument entered), command (2nd argument) and args (remaining arguments)
	directive := input[0]
	command := ""
	if len(input) > 1 {
		command = string(input[1])
	}
	var args []string
	if len(input) > 2 {
		args = input[2:]
	}

	log.Printf("Directive: %s, Command: %s, Args: %v", directive, command, args)

	switch string(input[0]) {
	case commandStart:
		if len(input) < 2 {
			return nil, fmt.Errorf("start command must contain the name of a job to execute on the host")
		}
		cmd := input[1]

		if len(input) >= 3 {
			args = input[2:]
		}

		return NewStartCommand(client, cmd, args...), nil
	case commandStop:
		jobIdStr := input[1]

		var jobId uuid.UUID
		if jobId, err = uuid.Parse(jobIdStr); err != nil {
			return nil, fmt.Errorf("job ID must be a valid UUID")
		}

		return NewStopCommand(client, jobId), nil
	case commandStatus:
		jobIdStr := input[1]

		var jobId uuid.UUID
		if jobId, err = uuid.Parse(jobIdStr); err != nil {
			return nil, fmt.Errorf("job ID must be a valid UUID")
		}

		return NewStatusCommand(client, jobId), nil
	case commandStream:
		jobIdStr := input[1]

		var jobId uuid.UUID
		if jobId, err = uuid.Parse(jobIdStr); err != nil {
			return nil, fmt.Errorf("job ID must be a valid UUID")
		}

		return NewStreamCommand(client, jobId), nil
	default:
		return nil, fmt.Errorf("invalid command: %s", input[0])
	}
}
