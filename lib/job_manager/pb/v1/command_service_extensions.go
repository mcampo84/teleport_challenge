package v1

import "github.com/google/uuid"

func (sr *StatusRequest) UuidFromUuid() uuid.UUID {
	return uuid.MustParse(sr.Uuid)
}

func (sr *StopRequest) UuidFromUuid() uuid.UUID {
	return uuid.MustParse(sr.Uuid)
}

func (sr *StreamOutputRequest) UuidFromUuid() uuid.UUID {
	return uuid.MustParse(sr.Uuid)
}
