package client

// Message represents a message from the message-broker.
type Message struct {
	// The type of the message: person.created , person.deleted , friendship.created
	Type string
	// The raw data of the message, encoded to json.
	Data []byte
}
