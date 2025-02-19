package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// 基础类型
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

func Stringp(key string, val *string) Field {
	return Field{Key: key, Value: val}
}

func Strings(key string, val []string) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Intp(key string, val *int) Field {
	return Field{Key: key, Value: val}
}

func Ints(key string, val []int) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Int64p(key string, val *int64) Field {
	return Field{Key: key, Value: val}
}

func Int64s(key string, val []int64) Field {
	return Field{Key: key, Value: val}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Value: val}
}

func Int32p(key string, val *int32) Field {
	return Field{Key: key, Value: val}
}

func Int32s(key string, val []int32) Field {
	return Field{Key: key, Value: val}
}

func Int16(key string, val int16) Field {
	return Field{Key: key, Value: val}
}

func Int16p(key string, val *int16) Field {
	return Field{Key: key, Value: val}
}

func Int16s(key string, val []int16) Field {
	return Field{Key: key, Value: val}
}

func Int8(key string, val int8) Field {
	return Field{Key: key, Value: val}
}

func Int8p(key string, val *int8) Field {
	return Field{Key: key, Value: val}
}

func Int8s(key string, val []int8) Field {
	return Field{Key: key, Value: val}
}

// Unsigned integers
func Uint(key string, val uint) Field {
	return Field{Key: key, Value: val}
}

func Uintp(key string, val *uint) Field {
	return Field{Key: key, Value: val}
}

func Uints(key string, val []uint) Field {
	return Field{Key: key, Value: val}
}

func Uint64(key string, val uint64) Field {
	return Field{Key: key, Value: val}
}

func Uint64p(key string, val *uint64) Field {
	return Field{Key: key, Value: val}
}

func Uint64s(key string, val []uint64) Field {
	return Field{Key: key, Value: val}
}

func Uint32(key string, val uint32) Field {
	return Field{Key: key, Value: val}
}

func Uint32p(key string, val *uint32) Field {
	return Field{Key: key, Value: val}
}

func Uint32s(key string, val []uint32) Field {
	return Field{Key: key, Value: val}
}

func Uint16(key string, val uint16) Field {
	return Field{Key: key, Value: val}
}

func Uint16p(key string, val *uint16) Field {
	return Field{Key: key, Value: val}
}

func Uint16s(key string, val []uint16) Field {
	return Field{Key: key, Value: val}
}

func Uint8(key string, val uint8) Field {
	return Field{Key: key, Value: val}
}

func Uint8p(key string, val *uint8) Field {
	return Field{Key: key, Value: val}
}

// 浮点数
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

func Float64p(key string, val *float64) Field {
	return Field{Key: key, Value: val}
}

func Float64s(key string, val []float64) Field {
	return Field{Key: key, Value: val}
}

func Float32(key string, val float32) Field {
	return Field{Key: key, Value: val}
}

func Float32p(key string, val *float32) Field {
	return Field{Key: key, Value: val}
}

func Float32s(key string, val []float32) Field {
	return Field{Key: key, Value: val}
}

// Complex numbers
func Complex128(key string, val complex128) Field {
	return Field{Key: key, Value: val}
}

func Complex64(key string, val complex64) Field {
	return Field{Key: key, Value: val}
}

// Bool types
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Boolp(key string, val *bool) Field {
	return Field{Key: key, Value: val}
}

func Bools(key string, val []bool) Field {
	return Field{Key: key, Value: val}
}

// Binary data
func Binary(key string, val []byte) Field {
	return Field{Key: key, Value: val}
}

// Time types
func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func Timep(key string, val *time.Time) Field {
	return Field{Key: key, Value: val}
}

func Times(key string, val []time.Time) Field {
	return Field{Key: key, Value: val}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

func Durationp(key string, val *time.Duration) Field {
	return Field{Key: key, Value: val}
}

func Durations(key string, val []time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Err types
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

func Errors(key string, errs []error) Field {
	return Field{Key: key, Value: errs}
}

// Special types
func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Value: val}
}

func Object(key string, val zapcore.ObjectMarshaler) Field {
	return Field{Key: key, Value: val}
}

func Array(key string, val zapcore.ArrayMarshaler) Field {
	return Field{Key: key, Value: val}
}

// Any handles any other type
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}

// 转换为 zap.Field
func toZapFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = toZapField(f)
	}
	return zapFields
}

// 单个字段转换
func toZapField(f Field) zap.Field {
	switch v := f.Value.(type) {
	case string:
		return zap.String(f.Key, v)
	case []string:
		return zap.Strings(f.Key, v)
	case *string:
		return zap.Stringp(f.Key, v)
	case int:
		return zap.Int(f.Key, v)
	case []int:
		return zap.Ints(f.Key, v)
	case *int:
		return zap.Intp(f.Key, v)
	case int64:
		return zap.Int64(f.Key, v)
	case []int64:
		return zap.Int64s(f.Key, v)
	case *int64:
		return zap.Int64p(f.Key, v)
	case int32:
		return zap.Int32(f.Key, v)
	case []int32:
		return zap.Int32s(f.Key, v)
	case *int32:
		return zap.Int32p(f.Key, v)
	case int16:
		return zap.Int16(f.Key, v)
	case []int16:
		return zap.Int16s(f.Key, v)
	case *int16:
		return zap.Int16p(f.Key, v)
	case int8:
		return zap.Int8(f.Key, v)
	case []int8:
		return zap.Int8s(f.Key, v)
	case *int8:
		return zap.Int8p(f.Key, v)
	case uint:
		return zap.Uint(f.Key, v)
	case []uint:
		return zap.Uints(f.Key, v)
	case *uint:
		return zap.Uintp(f.Key, v)
	case uint64:
		return zap.Uint64(f.Key, v)
	case []uint64:
		return zap.Uint64s(f.Key, v)
	case *uint64:
		return zap.Uint64p(f.Key, v)
	case uint32:
		return zap.Uint32(f.Key, v)
	case []uint32:
		return zap.Uint32s(f.Key, v)
	case *uint32:
		return zap.Uint32p(f.Key, v)
	case uint16:
		return zap.Uint16(f.Key, v)
	case []uint16:
		return zap.Uint16s(f.Key, v)
	case *uint16:
		return zap.Uint16p(f.Key, v)
	case uint8:
		return zap.Uint8(f.Key, v)
	case *uint8:
		return zap.Uint8p(f.Key, v)
	case float64:
		return zap.Float64(f.Key, v)
	case []float64:
		return zap.Float64s(f.Key, v)
	case *float64:
		return zap.Float64p(f.Key, v)
	case float32:
		return zap.Float32(f.Key, v)
	case []float32:
		return zap.Float32s(f.Key, v)
	case *float32:
		return zap.Float32p(f.Key, v)
	case complex128:
		return zap.Complex128(f.Key, v)
	case complex64:
		return zap.Complex64(f.Key, v)
	case bool:
		return zap.Bool(f.Key, v)
	case []bool:
		return zap.Bools(f.Key, v)
	case *bool:
		return zap.Boolp(f.Key, v)
	case []byte:
		return zap.Binary(f.Key, v)
	case time.Time:
		return zap.Time(f.Key, v)
	case []time.Time:
		return zap.Times(f.Key, v)
	case *time.Time:
		return zap.Timep(f.Key, v)
	case time.Duration:
		return zap.Duration(f.Key, v)
	case []time.Duration:
		return zap.Durations(f.Key, v)
	case *time.Duration:
		return zap.Durationp(f.Key, v)
	case error:
		return zap.Error(v)
	case []error:
		return zap.Errors(f.Key, v)
	case fmt.Stringer:
		return zap.Stringer(f.Key, v)
	case zapcore.ObjectMarshaler:
		return zap.Object(f.Key, v)
	case zapcore.ArrayMarshaler:
		return zap.Array(f.Key, v)
	default:
		return zap.Any(f.Key, v)
	}
}
