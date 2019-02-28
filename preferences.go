package cf

// #import <CoreFoundation/CoreFoundation.h>
import "C"
import (
	"fmt"

	"github.com/pkg/errors"
)

var PreferencesCurrentUser = "kCFPreferencesCurrentUser"
var PreferencesCurrentHost = "kCFPreferencesCurrentHost"
var PreferencesAnyHost = "kCFPreferencesAnyHost"

func Preferences(key, appID, userName, hostName string) (interface{}, error) {
	pool := Pool{}
	defer pool.Release()

	var key_, appID_, userName_, hostName_ stringRef
	var err error

	if key_, err = pool.String(key); err != nil {
		return nil, errors.Wrapf(err, "failed Preferences(%s)", key)
	}
	if appID_, err = pool.String(appID); err != nil {
		return nil, errors.Wrapf(err, "failed Preferences(%s)", key)
	}
	if userName_, err = pool.String(userName); err != nil {
		return nil, errors.Wrapf(err, "failed Preferences(%s)", key)
	}
	if hostName_, err = pool.String(hostName); err != nil {
		return nil, errors.Wrapf(err, "failed Preferences(%s)", key)
	}

	prefs := typeRef(C.CFPreferencesCopyValue(C.CFStringRef(key_), C.CFStringRef(appID_),
		C.CFStringRef(userName_), C.CFStringRef(hostName_)))
	defer Release(prefs)
	return prefs.Goize()
}

func PreferencesSet(key string, value interface{}, appID string, userName string, hostName string) error {
	pool := &Pool{}
	defer pool.Release()

	var key_, appID_, userName_, hostName_ stringRef
	var val_ typeRef
	var err error

	if key_, err = pool.String(key); err != nil {
		return errors.Wrapf(err, "failed PreferencesSet(%s)", key)
	}
	if appID_, err = pool.String(appID); err != nil {
		return errors.Wrapf(err, "failed PreferencesSet(%s)", key)
	}
	if userName_, err = pool.String(userName); err != nil {
		return errors.Wrapf(err, "failed PreferencesSet(%s)", key)
	}
	if hostName_, err = pool.String(hostName); err != nil {
		return errors.Wrapf(err, "failed PreferencesSet(%s)", key)
	}

	if val_, err = pool.Object(value); err != nil {
		return errors.Wrapf(err, "failed PreferencesSet(%s)", key)
	}
	C.CFPreferencesSetValue(C.CFStringRef(key_), C.CFTypeRef(val_), C.CFStringRef(appID_),
		C.CFStringRef(userName_), C.CFStringRef(hostName_))
	return nil
}

func PreferencesSetMulti(keys map[string]interface{}, appID string, userName string, hostName string) error {
	pool := &Pool{}
	defer pool.Release()

	cfAppID, err := pool.String(appID)
	if err != nil {
		return errors.Wrapf(err, "failed PreferencesSetMulti")
	}
	cfUserName, err := pool.String(userName)
	if err != nil {
		return errors.Wrap(err, "failed PreferercesSetMulti")
	}
	cfHostName, err := pool.String(hostName)
	if err != nil {
		return errors.Wrap(err, "failed PreferencesSetMulti")
	}

	delKeys := []string{}
	setKeys := map[string]interface{}{}
	for k, v := range keys {
		if v == nil {
			delKeys = append(delKeys, k)
		} else {
			setKeys[k] = v
		}
	}
	cfDelKeys, err := pool.Array(delKeys)
	if err != nil {
		return errors.Wrap(err, "failed PreferencesSetMulti")
	}
	cfSetKeys, err := pool.Dictionary(setKeys)
	if err != nil {
		return fmt.Errorf("Unable to convert %q to CFDictionary", setKeys)
	}
	C.CFPreferencesSetMultiple(C.CFDictionaryRef(cfSetKeys), C.CFArrayRef(cfDelKeys),
		C.CFStringRef(cfAppID), C.CFStringRef(cfUserName), C.CFStringRef(cfHostName))
	return nil
}

func PreferencesSynchronize(appID, userName, hostName string) (bool, error) {
	pool := &Pool{}
	defer pool.Release()

	var appID_, userName_, hostName_ stringRef
	var err error

	if appID_, err = pool.String(appID); err != nil {
		return false, errors.Wrap(err, "failed PreferencesSynchronize")
	}
	if userName_, err = pool.String(userName); err != nil {
		return false, errors.Wrap(err, "failed PreferencesSynchronize")
	}
	if hostName_, err = pool.String(hostName); err != nil {
		return false, errors.Wrap(err, "failed PreferencesSynchronize")
	}

	return C.CFPreferencesSynchronize(C.CFStringRef(appID_), C.CFStringRef(userName_),
		C.CFStringRef(hostName_)) != 0, nil
}
