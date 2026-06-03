package bson

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type ObjectID [12]byte

type M map[string]any

type D []E

type E struct {
	Key   string
	Value any
}

var NilObjectID ObjectID

func NewObjectID() ObjectID { var id ObjectID; _, _ = rand.Read(id[:]); return id }
func ObjectIDFromHex(s string) (ObjectID, error) {
	var id ObjectID
	if len(s) != 24 {
		return id, fmt.Errorf("the provided hex string is not a valid ObjectID")
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}
	copy(id[:], decoded)
	return id, nil
}
func IsObjectIDHex(s string) bool                { _, err := ObjectIDFromHex(s); return err == nil }
func (id ObjectID) IsZero() bool                 { return id == NilObjectID }
func (id ObjectID) Hex() string                  { return hex.EncodeToString(id[:]) }
func (id ObjectID) MarshalJSON() ([]byte, error) { return json.Marshal(id.Hex()) }
func (id *ObjectID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ObjectIDFromHex(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}
