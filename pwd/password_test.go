package pwd

import (
	"os"
	"testing"
)

var pwdStore *Store

func setup() {
	pwdStore = NewStore(64)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestHash_validPassword_returnsHashedValue(t *testing.T) {
	password := "angryMonkey"
	expected := "ZEHhWB65gUlzdVwtDQArEyx-KVLzp_aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A-gf7Q=="

	hashed := Hash(password)
	if hashed != expected {
		t.Errorf("unexpected hash. expected=%s, got=%s", expected, hashed)
	}
}

func TestHash_PutGetPassword_noError(t *testing.T) {
	pwd := "aPassword"
	password := &Password{
		Clear: "aPassword",
	}

	pwdStore.Put(1, *password)

	getPwd := pwdStore.Get(1)

	if getPwd.Clear != pwd {
		t.Errorf("unexpected password. expected=%s, got=%s", pwd, getPwd.Clear)
	}
}

func TestHash_GetNonExistingPassword_returnsEmptyPassword(t *testing.T) {
	getPwd := pwdStore.Get(2)

	if !IsEmpty(getPwd) {
		t.Errorf("expected password to be empty. got=%+v", getPwd)
	}
}

func TestHash_PasswordTooLong_returnsError(t *testing.T) {
	password66 := "angryMonkeyangryMonkeyangryMonkeyangryMonkeyangryMonkeyangryMonkey"

	if err := pwdStore.Validate(password66); err == nil {
		t.Errorf("a password exceeding maxLength must return an error")
	}
}

func TestHash_PasswordEqualsMaxLength_returnsNoError(t *testing.T) {
	password64 := "angryMonkeyangryMonkeyangryMonkeyangryMonkeyangryMonkeyangryMonk"

	if err := pwdStore.Validate(password64); err != nil {
		t.Errorf("a password mathing maxLength must not return an error. got error=%+v", err)
	}
}

