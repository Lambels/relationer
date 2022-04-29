package internal

import (
	"time"
)

type Person struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func (p *Person) Validate() error {
	if p.Name == "" {
		return Errorf(EINVALID, "name is a required field")
	}
	return nil
}
