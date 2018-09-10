package pwd

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"sync"
	"unicode/utf8"
)

type Store struct {
	pwdMap       map[uint64]Password
	mutex        *sync.Mutex
	maxPwdLength int
}

func NewStore(maxPwdLength int) *Store {
	s := &Store{
		pwdMap:       make(map[uint64]Password),
		mutex:        &sync.Mutex{},
		maxPwdLength: maxPwdLength,
	}
	return s
}

type Password struct {
	Clear  string
	Hashed string
}

func (s *Store) Put(id uint64, pwd Password) {
	s.mutex.Lock()
	s.pwdMap[id] = pwd
	s.mutex.Unlock()
}

func (s *Store) Get(id uint64) Password {
	var password Password
	s.mutex.Lock()
	password = s.pwdMap[id]
	s.mutex.Unlock()
	return password
}

func (s *Store) Validate(clearText string) error {
	strLen := int(utf8.RuneCount([]byte(clearText)))
	if strLen > s.maxPwdLength {
		return fmt.Errorf("password exceeded max length of '%d'", s.maxPwdLength)
	}
	return nil
}

func Hash(clearText string) string {
	hasher := sha512.New()
	hasher.Write([]byte(clearText))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func IsEmpty(pwd Password) bool {
	return pwd.Clear == "" && pwd.Hashed == ""
}
