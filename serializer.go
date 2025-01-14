package sessions

import "encoding/json"

type Serializer struct{}

func (js *Serializer) Serialize(session *Session) ([]byte, error) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	return json.Marshal(session.data)
}

func (js *Serializer) Deserialize(data []byte, session *Session) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.data == nil {
		session.data = &sessionData{}
	}

	return json.Unmarshal(data, session.data)
}
