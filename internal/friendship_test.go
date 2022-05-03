package internal

import "testing"

func TestFriendshipValidate(t *testing.T) {
	t.Run("No Err", func(t *testing.T) {
		f := Friendship{
			P1:   &Person{Name: "A"},
			With: make([]*Person, 1),
		}

		f.With = append(f.With, &Person{
			Name: "B",
		})

		if err := f.Validate(); err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}
	})

	t.Run("Err", func(t *testing.T) {
		t.Run("Invalid Person", func(t *testing.T) {
			f := Friendship{
				P1:   &Person{},
				With: make([]*Person, 1),
			}

			f.With = append(f.With, &Person{
				Name: "B",
			})

			if err := f.Validate().(*Error); err.Code() != EINVALID {
				t.Fatalf("expected EINVALID error but got: %v", err)
			}
		})

		t.Run("Nil Person", func(t *testing.T) {
			f := Friendship{
				With: make([]*Person, 1),
			}

			f.With = append(f.With, &Person{
				Name: "B",
			})

			if err := f.Validate().(*Error); err.Code() != EINVALID {
				t.Fatalf("expected EINVALID error but got: %v", err)
			}
		})

		t.Run("Invalid Slice", func(t *testing.T) {
			f := Friendship{
				P1:   &Person{Name: "A"},
				With: make([]*Person, 0),
			}

			if err := f.Validate().(*Error); err.Code() != EINVALID {
				t.Fatalf("expected EINVALID error but got: %v", err)
			}
		})
	})
}
