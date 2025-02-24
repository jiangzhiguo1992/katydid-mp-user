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
	Value any
}

func FString(key string, val string) Field {
	return Field{Key: key, Value: val}
}

func FStringp(key string, val *string) Field {
	return Field{Key: key, Value: val}
}

func FStrings(key string, val []string) Field {
	return Field{Key: key, Value: val}
}

func FInt(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func FIntp(key string, val *int) Field {
	return Field{Key: key, Value: val}
}

func FInts(key string, val []int) Field {
	return Field{Key: key, Value: val}
}

func FInt64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func FInt64p(key string, val *int64) Field {
	return Field{Key: key, Value: val}
}

func FInt64s(key string, val []int64) Field {
	return Field{Key: key, Value: val}
}

func FInt32(key string, val int32) Field {
	return Field{Key: key, Value: val}
}

func FInt32p(key string, val *int32) Field {
	return Field{Key: key, Value: val}
}

func FInt32s(key string, val []int32) Field {
	return Field{Key: key, Value: val}
}

func FInt16(key string, val int16) Field {
	return Field{Key: key, Value: val}
}

func FInt16p(key string, val *int16) Field {
	return Field{Key: key, Value: val}
}

func FInt16s(key string, val []int16) Field {
	return Field{Key: key, Value: val}
}

func FInt8(key string, val int8) Field {
	return Field{Key: key, Value: val}
}

func FInt8p(key string, val *int8) Field {
	return Field{Key: key, Value: val}
}

func FInt8s(key string, val []int8) Field {
	return Field{Key: key, Value: val}
}

func FUint(key string, val uint) Field {
	return Field{Key: key, Value: val}
}

func FUintp(key string, val *uint) Field {
	return Field{Key: key, Value: val}
}

func FUints(key string, val []uint) Field {
	return Field{Key: key, Value: val}
}

func FUint64(key string, val uint64) Field {
	return Field{Key: key, Value: val}
}

func FUint64p(key string, val *uint64) Field {
	return Field{Key: key, Value: val}
}

func FUint64s(key string, val []uint64) Field {
	return Field{Key: key, Value: val}
}

func FUint32(key string, val uint32) Field {
	return Field{Key: key, Value: val}
}

func FUint32p(key string, val *uint32) Field {
	return Field{Key: key, Value: val}
}

func FUint32s(key string, val []uint32) Field {
	return Field{Key: key, Value: val}
}

func FUint16(key string, val uint16) Field {
	return Field{Key: key, Value: val}
}

func FUint16p(key string, val *uint16) Field {
	return Field{Key: key, Value: val}
}

func FUint16s(key string, val []uint16) Field {
	return Field{Key: key, Value: val}
}

func FUint8(key string, val uint8) Field {
	return Field{Key: key, Value: val}
}

func FUint8p(key string, val *uint8) Field {
	return Field{Key: key, Value: val}
}

func FFloat64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

func FFloat64p(key string, val *float64) Field {
	return Field{Key: key, Value: val}
}

func FFloat64s(key string, val []float64) Field {
	return Field{Key: key, Value: val}
}

func FFloat32(key string, val float32) Field {
	return Field{Key: key, Value: val}
}

func FFloat32p(key string, val *float32) Field {
	return Field{Key: key, Value: val}
}

func FFloat32s(key string, val []float32) Field {
	return Field{Key: key, Value: val}
}

func FComplex128(key string, val complex128) Field {
	return Field{Key: key, Value: val}
}

func FComplex64(key string, val complex64) Field {
	return Field{Key: key, Value: val}
}

func FBool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func FBoolp(key string, val *bool) Field {
	return Field{Key: key, Value: val}
}

func FBools(key string, val []bool) Field {
	return Field{Key: key, Value: val}
}

func FBinary(key string, val []byte) Field {
	return Field{Key: key, Value: val}
}

func FTime(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func FTimep(key string, val *time.Time) Field {
	return Field{Key: key, Value: val}
}

func FTimes(key string, val []time.Time) Field {
	return Field{Key: key, Value: val}
}

func FDuration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

func FDurationp(key string, val *time.Duration) Field {
	return Field{Key: key, Value: val}
}

func FDurations(key string, val []time.Duration) Field {
	return Field{Key: key, Value: val}
}

func FError(err error) Field {
	return Field{Key: "error", Value: err}
}

func FErrors(key string, errs []error) Field {
	return Field{Key: key, Value: errs}
}

func FStringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Value: val}
}

func FObject(key string, val zapcore.ObjectMarshaler) Field {
	return Field{Key: key, Value: val}
}

func FArray(key string, val zapcore.ArrayMarshaler) Field {
	return Field{Key: key, Value: val}
}

func FAny(key string, val any) Field {
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
