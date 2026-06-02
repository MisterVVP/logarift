package bson

import (
	"crypto/rand"
	"encoding/hex"
)

type ObjectID [12]byte

type M map[string]any

type D []E

type E struct {
	Key   string
	Value any
}

var NilObjectID ObjectID

func NewObjectID() ObjectID      { var id ObjectID; _, _ = rand.Read(id[:]); return id }
func (id ObjectID) IsZero() bool { return id == NilObjectID }
func (id ObjectID) Hex() string  { return hex.EncodeToString(id[:]) }
