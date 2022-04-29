package internal

// Friendship represents a friendship from the perspective of P1.
type Friendship struct {
	P1   *Person   `json:"p1"`
	With []*Person `json:"with"`
}

func (f Friendship) Validate() error {
	if f.P1 == nil {
		return Errorf(EINVALID, "nil pointer")
	}

	if err := f.P1.Validate(); err != nil {
		return WrapError(err, EINVALID, "friendship required a valid person")
	}

	if len(f.With) == 0 {
		return Errorf(EINVALID, "at least one person is required")
	}
	return nil
}
