package db

import (
	"RoleKeeper/cons"

	"github.com/sasha-s/go-deadlock"
)

type RoleData struct {
	Name string `json:",omitempty"`
	ID   uint64 `json:",omitempty"`
}

type GuildData struct {
	//Name type bytes
	LID      uint32     `json:",omitempty"`
	Customer uint64     `json:",omitempty"`
	Guild    uint64     `json:"-"` //Already in JSON as KEY
	Added    uint32     `json:",omitempty"`
	Modified uint32     `json:",omitempty"`
	Donator  uint8      `json:",omitempty"`
	Roles    []RoleData `json:",omitempty"`

	/* Not on disk */
	Lock deadlock.RWMutex `json:"-"`
}

var (
	LID_TOP         uint32 = 0
	GuildLookup     map[uint64]*GuildData
	GuildLookupLock deadlock.RWMutex
	ThreadCount     int

	Database [cons.MaxGuilds]*GuildData
)
