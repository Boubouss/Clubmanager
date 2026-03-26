package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {

  tests := []struct{
    name string
    data map[string]string
    wantErr bool
  }{
    {"Valid user", map[string]string{
      "username": "rbouselh",
      "email": "test@mail.com",
      "phonenumber": "0606060606",
      "password": "Password123!",
    }, false},
    {"Blank username", map[string]string{
      "username": "",
      "email": "test@mail.com",
      "phonenumber": "0606060606",
      "password": "Password123!",
    }, true},
    {"Invalid email", map[string]string{
      "username": "rbouselh",
      "email": "testmail.com",
      "phonenumber": "0606060606",
      "password": "Password123!",
    }, true},
    {"Invalid phonenumber", map[string]string{
      "username": "rbouselh",
      "email": "test@mail.com",
      "phonenumber": "606060606",
      "password": "Password123!",
    }, true},
    {"Invalid password", map[string]string{
      "username": "rbouselh",
      "email": "test@mail.com",
      "phonenumber": "0606060606",
      "password": "password",
    }, true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func (t *testing.T) {
      _, errs := NewUser(tt.data)
      if tt.wantErr {
        assert.True(t, len(errs) > 0)
      } else {
        assert.True(t, len(errs) == 0)
      }
    })
  }
}

func TestIsEmail(t *testing.T) {
  tests := []struct{
    name string
    data string
    wantErr bool
  }{
    {"Valid email", "test@email.com", false},
    {"Without @", "testemail.com", true},
    {"Without .", "test@emailcom", true},
    {"Without last part", "test@email.", true},
    {"Without center part", "test@.com", true},
    {"Without first part", "@email.com", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      ok, _ := IsEmail(tt.data)
      if tt.wantErr {
        assert.False(t, ok)
      } else {
        assert.True(t, ok)
      }
    })
  }
}

func TestIsPhoneNumber(t *testing.T) {
  tests := []struct{
    name string
    data string
    wantErr bool
  }{
    {"Valid phonenumber", "0606060606", false},
    {"Fist number not 0", "6060606060", true},
    {"Wrong size", "06", true},
    {"Blank phonenumber", "", true},
    {"Letter in phonenumber", "060606060r", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      ok, _ := IsPhoneNumber(tt.data)
      if tt.wantErr {
        assert.False(t, ok)
      } else {
        assert.True(t, ok)
      }
    })
  }
}

func TestIsValidUsername(t *testing.T) {
  tests := []struct{
    name string
    data string
    wantErr bool
  }{
    {"Valid username", "rbouselh", false},
    {"Invalid char", "rbouselh!", true},
    {"Blank username", "", true},
    {"Too short", "r", true},
    {"Too long", "rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      ok, _ := IsValidUsername(tt.data)
      if tt.wantErr {
        assert.False(t, ok)
      } else {
        assert.True(t, ok)
      }
    })
  }
}

func TestIsValidPassword(t *testing.T) {
  tests := []struct{
    name string
    data string
    wantErr bool
  }{
    {"Valid username", "Password123!", false},
    {"Without special", "Password123", true},
    {"Without upper", "password123!", true},
    {"Without lower", "PASSWORD123!", true},
    {"Without num", "Password!", true},
    {"Too short", "aA1!", true},
    {"Too long", "A1!rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      ok, _ := IsValidPassword(tt.data)
      if tt.wantErr {
        assert.False(t, ok)
      } else {
        assert.True(t, ok)
      }
    })
  }
}

