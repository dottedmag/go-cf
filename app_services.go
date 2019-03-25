package cf

// #include <ApplicationServices/ApplicationServices.h>
// #include <objc/objc.h>
//
// typedef void *CGSConnection;
// CGSConnection _CGSDefaultConnection(void);
// void CGSSetSwipeScrollDirection(const CGSConnection cid, BOOL dir);
//
// void CoreDockGetOrientationAndPinning(int *orientation, int *pinning);
// void CoreDockSetOrientationAndPinning(int orientation, int pinning);
// void CoreDockSetTileSize(float size);
// float CoreDockGetTileSize();
//
// int CoreDockCopyPreferences(CFArrayRef keys, CFPropertyListRef *r);
// void CoreDockSetPreferences(CFPropertyListRef r);
//
// #cgo LDFLAGS: -framework ApplicationServices
import "C"
import (
	"fmt"

	"github.com/pkg/errors"
)

func CGSetSwipeScrollDirection(dir bool) error {
	var byteDir byte
	if dir {
		byteDir = 1
	}
	conn := C._CGSDefaultConnection()
	if conn == nil {
		return errors.New("failed to establesh connection to WindowServer")
	}
	C.CGSSetSwipeScrollDirection(conn, C.schar(byteDir))
	return nil
}

func CoreDockCopyPreferences(keys []string) (map[string]interface{}, error) {
	pool := Pool{}
	defer pool.Release()

	arr, err := pool.Array(keys)
	if err != nil {
		return nil, err
	}

	var p C.CFPropertyListRef
	cErr := int(C.CoreDockCopyPreferences(C.CFArrayRef(arr), &p))
	if cErr == 0 {
		v, err := typeRef(p).Goize()
		if err != nil {
			return nil, err
		}
		return v.(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("CoreDockCopyPreferences error %d", cErr)
}

func CoreDockSetPreferences(values map[string]interface{}) error {
	pool := &Pool{}
	defer pool.Release()

	var values_ typeRef
	var err error
	if values_, err = pool.Object(values); err != nil {
		return errors.Wrapf(err, "failed CoreDockSetPreferences(%v)", values)
	}
	C.CoreDockSetPreferences(C.CFTypeRef(values_))
	return nil
}

func CoreDockGetOrientationAndPinning() (orientation int, pinning int) {
	var ori, pi C.int
	C.CoreDockGetOrientationAndPinning(&ori, &pi)
	return int(ori), int(pi)
}

func CoreDockSetOrientationAndPinning(orientation, pinning int) {
	C.CoreDockSetOrientationAndPinning(C.int(orientation), C.int(pinning))
}

func CoreDockGetTileSize() float64 {
	return float64(C.CoreDockGetTileSize())
}

func CoreDockSetTileSize(size float64) {
	C.CoreDockSetTileSize(C.float(size))
}
