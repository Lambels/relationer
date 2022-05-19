package internal

// Friendship represents a friendship from the perspective of P1.
type Friendship struct {
	P1   *Person `json:"p1"`
	With []int64 `json:"with"`
}

func (f Friendship) Validate() error {
	if f.P1 == nil {
		return Errorf(EINVALID, "nil pointer")
	}

	if len(f.With) == 0 {
		return Errorf(EINVALID, "at least one person is required")
	}
	return nil
}
