package colonio

/*
#cgo CFLAGS: -I${SRCDIR}/.
#cgo LDFLAGS: -L${SRCDIR}/../colonio/output -L${SRCDIR}/../colonio/output/lib -lcolonio -lwebrtc -lm -lprotobuf -lstdc++ bridge.a
#cgo pkg-config: openssl

#include "../colonio/src/colonio/colonio.h"

extern const unsigned int cgo_colonio_nid_length;

// colonio
colonio_error_t *cgo_colonio_connect(colonio_t *colonio, _GoString_ url, _GoString_ token);
colonio_map_t cgo_colonio_access_map(colonio_t *colonio, _GoString_ name);
colonio_pubsub_2d_t cgo_colonio_access_pubsub_2d(colonio_t *colonio, _GoString_ name);

// value
void cgo_colonio_value_set_string(colonio_value_t *value, _GoString_ s);

// pubsub
colonio_error_t *cgo_colonio_pubsub_2d_publish(
    colonio_pubsub_2d_t *pubsub_2d, _GoString_ name, double x, double y, double r, const colonio_value_t *value,
    uint32_t opt);
void cgo_cb_colonio_pubsub_2d_on(colonio_pubsub_2d_t *pubsub_2d, void *ptr, const colonio_value_t *val);
void cgo_colonio_pubsub_2d_on(colonio_pubsub_2d_t *pubsub_2d, _GoString_ name, void *ptr);
void cgo_colonio_pubsub_2d_off(colonio_pubsub_2d_t *pubsub_2d, _GoString_ name);
*/
import "C"

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// Colonio is an instance. It is equivalent to one node.
type Colonio struct {
	cInstance     C.struct_colonio_s
	mapCache      map[string]*Map
	pubsub2DCache map[string]*Pubsub2D
}

// Value is an instance, it is equivalent to one value.
type Value struct {
	valueType C.enum_COLONIO_VALUE_TYPE
	vBool     bool
	vInt      int64
	vDouble   float64
	vString   string
}

type Map struct {
	cInstance C.struct_colonio_map_s
}

type Pubsub2D struct {
	cInstance C.struct_colonio_pubsub_2d_s
	cbMutex   sync.RWMutex
	cbMap     map[*string]func(*Value)
}

var pubsub2DMutex sync.RWMutex
var pubsub2DMap map[*C.struct_colonio_pubsub_2d_s]*Pubsub2D

func init() {
	pubsub2DMutex = sync.RWMutex{}
	pubsub2DMap = make(map[*C.struct_colonio_pubsub_2d_s]*Pubsub2D)
}

func convertError(err *C.struct_colonio_error_s) error {
	return fmt.Errorf("colonio error")
}

// NewColonio creates a new initialized instance.
func NewColonio() (*Colonio, error) {
	instance := &Colonio{}
	err := C.colonio_init(&instance.cInstance)
	if err != nil {
		return nil, convertError(err)
	}
	return instance, nil
}

// Connect to seed and join the cluster.
func (c *Colonio) Connect(url, token string) error {
	err := C.cgo_colonio_connect(&c.cInstance, url, token)
	if err != nil {
		return convertError(err)
	}
	return nil
}

// Disconnect from the cluster and the seed.
func (c *Colonio) Disconnect() error {
	err := C.colonio_disconnect(&c.cInstance)
	if err != nil {
		return convertError(err)
	}
	return nil
}

func (c *Colonio) AccessMap(name string) *Map {
	if ret, ok := c.mapCache[name]; ok {
		return ret
	}

	instance := &Map{
		cInstance: C.cgo_colonio_access_map(&c.cInstance, name),
	}

	c.mapCache[name] = instance
	return instance
}

func (c *Colonio) AccessPubsub2D(name string) *Pubsub2D {
	if ret, ok := c.pubsub2DCache[name]; ok {
		return ret
	}

	instance := &Pubsub2D{
		cInstance: C.cgo_colonio_access_pubsub_2d(&c.cInstance, name),
	}
	pubsub2DMutex.Lock()
	defer pubsub2DMutex.Unlock()
	if _, ok := pubsub2DMap[&instance.cInstance]; ok {

	} else {
		pubsub2DMap[&instance.cInstance] = instance
	}

	c.pubsub2DCache[name] = instance
	return instance
}

func (c *Colonio) GetLocalNid() string {
	buf := make([]byte, C.cgo_colonio_nid_length+1)
	data := (*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data
	C.colonio_get_local_nid(&c.cInstance, (*C.char)(unsafe.Pointer(data)), nil)
	return string(buf)
}

func (c *Colonio) SetPosition(x, y float64) (float64, float64, error) {
	cX := C.double(x)
	cY := C.double(y)
	err := C.colonio_set_position(&c.cInstance, &cX, &cY)
	if err != nil {
		return 0, 0, convertError(err)
	}
	return float64(cX), float64(cY), nil
}

// Quit is the finalizer of the instance.
func (c *Colonio) Quit() error {
	err := C.colonio_quit(&c.cInstance)
	if err != nil {
		return convertError(err)
	}
	return nil
}

func newValue(cValue *C.struct_colonio_value_s) *Value {
	valueType := C.enum_COLONIO_VALUE_TYPE(C.colonio_value_get_type(cValue))
	switch valueType {
	case C.COLONIO_VALUE_TYPE_BOOL:
		return &Value{
			valueType: valueType,
			vBool:     bool(C.colonio_value_get_bool(cValue)),
		}

	case C.COLONIO_VALUE_TYPE_INT:
		return &Value{
			valueType: valueType,
			vInt:      int64(C.colonio_value_get_int(cValue)),
		}

	case C.COLONIO_VALUE_TYPE_DOUBLE:
		return &Value{
			valueType: valueType,
			vDouble:   float64(C.colonio_value_get_double(cValue)),
		}

	case C.COLONIO_VALUE_TYPE_STRING:
		buf := make([]byte, uint(C.colonio_value_get_string_siz(cValue)))
		data := (*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data
		C.colonio_value_get_string(cValue, (*C.char)(unsafe.Pointer(data)))
		return &Value{
			valueType: valueType,
			vString:   string(buf),
		}

	default:
		return &Value{
			valueType: valueType,
		}
	}
}

func NewValue(v interface{}) (*Value, error) {
	val := &Value{}
	err := val.Set(v)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (v *Value) IsNil() bool {
	return v.valueType == C.COLONIO_VALUE_TYPE_NULL
}

func (v *Value) IsBool() bool {
	return v.valueType == C.COLONIO_VALUE_TYPE_BOOL
}

func (v *Value) IsInt() bool {
	return v.valueType == C.COLONIO_VALUE_TYPE_INT
}

func (v *Value) IsDouble() bool {
	return v.valueType == C.COLONIO_VALUE_TYPE_DOUBLE
}

func (v *Value) IsString() bool {
	return v.valueType == C.COLONIO_VALUE_TYPE_STRING
}

func (v *Value) Set(val interface{}) error {
	v.vString = ""

	if reflect.ValueOf(v).IsNil() {
		v.valueType = C.COLONIO_VALUE_TYPE_NULL
		return nil
	}

	switch val.(type) {
	case bool:
		v.valueType = C.COLONIO_VALUE_TYPE_BOOL
		v.vBool = val.(bool)
		return nil

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32:
		v.valueType = C.COLONIO_VALUE_TYPE_INT
		v.vInt = val.(int64)
		return nil

	case float32, float64:
		v.valueType = C.COLONIO_VALUE_TYPE_DOUBLE
		v.vDouble = val.(float64)
		return nil

	case string:
		v.valueType = C.COLONIO_VALUE_TYPE_STRING
		v.vString = val.(string)
		return nil
	}

	return fmt.Errorf("unsupported value type")
}

func (v *Value) GetBool() (bool, error) {
	if v.valueType != C.COLONIO_VALUE_TYPE_BOOL {
		return false, fmt.Errorf("the type of value is wrong")
	}
	return v.vBool, nil
}

func (v *Value) GetInt() (int64, error) {
	if v.valueType != C.COLONIO_VALUE_TYPE_INT {
		return 0, fmt.Errorf("the type of value is wrong")
	}
	return v.vInt, nil
}

func (v *Value) GetDouble() (float64, error) {
	if v.valueType != C.COLONIO_VALUE_TYPE_DOUBLE {
		return 0, fmt.Errorf("the type of value is wrong")
	}
	return v.vDouble, nil
}

func (v *Value) GetString() (string, error) {
	if v.valueType != C.COLONIO_VALUE_TYPE_STRING {
		return "", fmt.Errorf("the type of value is wrong")
	}
	return v.vString, nil
}

func (v *Value) writeOut(cValue *C.struct_colonio_value_s) {
	switch v.valueType {
	case C.COLONIO_VALUE_TYPE_BOOL:
		C.colonio_value_set_bool(cValue, C.bool(v.vBool))
	case C.COLONIO_VALUE_TYPE_INT:
		C.colonio_value_set_int(cValue, C.int64_t(v.vInt))
	case C.COLONIO_VALUE_TYPE_DOUBLE:
		C.colonio_value_set_double(cValue, C.double(v.vDouble))
	case C.COLONIO_VALUE_TYPE_STRING:
		C.cgo_colonio_value_set_string(cValue, v.vString)
	default:
		C.colonio_value_free(cValue)
	}
}

func (m *Map) Get(key interface{}) (*Value, error) {
	// key
	vKey, err := NewValue(key)
	if err != nil {
		return nil, err
	}
	cKey := C.struct_colonio_value_s{}
	C.colonio_value_init(&cKey)
	defer C.colonio_value_free(&cKey)
	vKey.writeOut(&cKey)

	// value
	cVal := C.struct_colonio_value_s{}
	C.colonio_value_init(&cVal)
	defer C.colonio_value_free(&cVal)

	// get
	cErr := C.colonio_map_get(&m.cInstance, &cKey, &cVal)
	if cErr != nil {
		return nil, convertError(cErr)
	}

	return newValue(&cVal), nil
}

func (m *Map) Set(key, val interface{}, opt uint32) error {
	// key
	vKey, err := NewValue(key)
	if err != nil {
		return err
	}
	cKey := C.struct_colonio_value_s{}
	C.colonio_value_init(&cKey)
	defer C.colonio_value_free(&cKey)
	vKey.writeOut(&cKey)

	// value
	vValue, err := NewValue(val)
	if err != nil {
		return err
	}
	cVal := C.struct_colonio_value_s{}
	C.colonio_value_init(&cVal)
	defer C.colonio_value_free(&cVal)
	vValue.writeOut(&cVal)

	// set
	cErr := C.colonio_map_set(&m.cInstance, &cKey, &cVal, C.uint32_t(opt))
	if cErr != nil {
		return convertError(cErr)
	}
	return nil
}

func (p *Pubsub2D) Publish(name string, x, y, r float64, val interface{}, opt uint32) error {
	// value
	vValue, err := NewValue(val)
	if err != nil {
		return err
	}
	cVal := C.struct_colonio_value_s{}
	C.colonio_value_init(&cVal)
	defer C.colonio_value_free(&cVal)
	vValue.writeOut(&cVal)

	// publish
	cErr := C.cgo_colonio_pubsub_2d_publish(&p.cInstance, name, C.double(x), C.double(y), C.double(r), &cVal, C.uint32_t(opt))
	if cErr != nil {
		return convertError(cErr)
	}
	return nil
}

//export cgoCbPubsub2DOn
func cgoCbPubsub2DOn(cInstancePtr *C.struct_colonio_pubsub_2d_s, ptr unsafe.Pointer, cVal *C.struct_colonio_value_s) {
	var ps2 *Pubsub2D
	{
		pubsub2DMutex.RLock()
		defer pubsub2DMutex.RUnlock()
		if p, ok := pubsub2DMap[cInstancePtr]; ok {
			ps2 = p
		} else {
			return
		}
	}

	var cb func(*Value)
	{
		ps2.cbMutex.RLock()
		defer ps2.cbMutex.RUnlock()
		if c, ok := ps2.cbMap[(*string)(ptr)]; ok {
			cb = c
		} else {
			return
		}
	}
	value := newValue(cVal)
	cb(value)
}

func (p *Pubsub2D) On(name string, cb func(*Value)) {
	p.cbMutex.Lock()
	defer p.cbMutex.Unlock()
	p.cbMap[&name] = cb
	C.cgo_colonio_pubsub_2d_on(&p.cInstance, name, unsafe.Pointer(&name))
}

func (p *Pubsub2D) Off(name string) {
	C.cgo_colonio_pubsub_2d_off(&p.cInstance, name)

	p.cbMutex.Lock()
	defer p.cbMutex.Unlock()

	for s, _ := range p.cbMap {
		if *s == name {
			delete(p.cbMap, s)
			return
		}
	}
}
