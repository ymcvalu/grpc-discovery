package instance

import (
	"encoding/json"
)

type Metadata map[string]string

type MetadataConvert func(Metadata) interface{}

type Instance struct {
	Env      string
	AppID    string
	Addr     string
	Port     int
	Metadata Metadata
}

func (m Metadata) ToMap() map[string]string {
	return map[string]string(m )
}

func (inst *Instance) Encode() string {
	byts, _ := json.Marshal(inst)
	return string(byts)
}

func (inst *Instance) Decode(byts []byte) {
	json.Unmarshal(byts, inst)
}
