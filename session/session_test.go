package mi_session

import "testing"

func TestInvalidSeparator(t *testing.T) {
	session := Session{"secret_key"}

	_, err := session.Unsign("abc")
	if err != nil {
		t.FailNow()
	}
}

func TestUnsign(t *testing.T) {
	session := Session{"~ew81z@dfHsaasa81<asd1!2>askjbb"}

	key, err := session.Unsign("f4987115-104b-47ae-8465-54c31f1e7848.gKbUzlueAlSFEw1J1wfNtCCHsF8")
	if err != nil {
		t.Fatal(err)
	}

	if key != "f4987115-104b-47ae-8465-54c31f1e7848" {
		t.Fatal("Expected f4987115-104b-47ae-8465-54c31f1e7848, got ", key)
	}
}
