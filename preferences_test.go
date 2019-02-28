package cf

import (
	"testing"
)

func TestSaveRestore(t *testing.T) {
	err := PreferencesSet("lala", "zuzu", "bobo", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	v, err := Preferences("lala", "bobo", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	if v != "zuzu" {
		panic("expected zuzu")
	}
	err = PreferencesSet("lala", nil, "bobo", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	v, err = Preferences("lala", "bobo", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	if v != nil {
		panic("unexpected lala")
	}
}

func TestMulti(t *testing.T) {
	err := PreferencesSet("a", "aval", "raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	err = PreferencesSet("b", "bval", "raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	err = PreferencesSetMulti(map[string]interface{}{"a": "aval2", "b": nil, "c": "cval"},
		"raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	v, err := Preferences("a", "raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	if v != "aval2" {
		panic("wrong a")
	}
	v, err = Preferences("b", "raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	if v != nil {
		panic("unexpected b")
	}
	v, err = Preferences("c", "raar", PreferencesCurrentUser, PreferencesAnyHost)
	if err != nil {
		panic(err)
	}
	if v != "cval" {
		panic("wrong c")
	}
}
