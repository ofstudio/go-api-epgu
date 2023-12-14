package utils

import "testing"

func TestGUID(t *testing.T) {
	t.Run("length", func(t *testing.T) {
		guid, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		if len(guid) != 36 {
			t.Errorf("expected length 36, got %d", len(guid))
		}
	})
	t.Run("format", func(t *testing.T) {
		guid, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		if guid[8] != '-' || guid[13] != '-' || guid[18] != '-' || guid[23] != '-' {
			t.Error("expected dashes at positions 8, 13, 18 and 23")
		}
	})
	t.Run("version", func(t *testing.T) {
		guid, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		if guid[14] != '4' {
			t.Error("expected version 4")
		}
	})
	t.Run("variant", func(t *testing.T) {
		guid, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		if guid[19] != '8' && guid[19] != '9' && guid[19] != 'a' && guid[19] != 'b' {
			t.Error("expected variant 8, 9, a or b")
		}
	})
	t.Run("uniqueness", func(t *testing.T) {
		guid1, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		guid2, err := GUID()
		if err != nil {
			t.Fatal(err)
		}
		if guid1 == guid2 {
			t.Error("expected unique GUIDs")
		}
	})
}
