package core

import (
	"github.com/M-O-S-E-S/mgm/mgm"
	"github.com/satori/go.uuid"
)

// LoadIarJob is the data field for jobs that are of type load_iar
type LoadIarJob struct {
	InventoryPath string
	Filename      uuid.UUID
	Status        string
}

// Identity is a simiangrid credential record
type Identity struct {
	Identifier string
	Credential string
	Type       string
	UserID     uuid.UUID
	Enabled    bool
}

type sessionLookup struct {
	jobLink      chan mgm.Job
	hostStatLink chan mgm.HostStat
	hostLink     chan mgm.Host
	accessLevel  uint8
}
