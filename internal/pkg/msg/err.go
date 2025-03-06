package msg

const (
	ErrCodeUnknown = 0
	ErrCodeDB      = 1000
)

const (
	ErrIdDBPkDuplicated   = "err_db_pk_duplicated"
	ErrIdDBAddNil         = "err_db_add_nil"
	ErrIdDBDelNil         = "err_db_del_nil"
	ErrIdDBUpdNil         = "err_db_upd_nil"
	ErrIdDBQueNil         = "err_db_que_nil"
	ErrIdDBFieldNil       = "err_db_field_nil"
	ErrIdDBFieldLarge     = "err_db_field_large"
	ErrIdDBFieldShort     = "err_db_field_short"
	ErrIdDBFieldMax       = "err_db_field_max"
	ErrIdDBFieldMin       = "err_db_field_min"
	ErrIdDBFieldRange     = "err_db_field_range"
	ErrIdDBFieldUnDefined = "err_db_field_undefined"
	ErrIdDBQueParams      = "err_db_que_params"
	ErrIdDBQueNone        = "err_db_que_none"
	ErrIdDBQueForeignNone = "err_db_que_foreign_none"
)

var (
	// ErrCodePatterns 错误信息映射
	ErrCodePatterns = map[int][]string{
		ErrCodeDB: {
			ErrIdDBPkDuplicated,
			ErrIdDBAddNil,
			ErrIdDBDelNil,
			ErrIdDBUpdNil,
			ErrIdDBQueNil,
			ErrIdDBFieldNil,
			ErrIdDBFieldLarge,
			ErrIdDBFieldShort,
			ErrIdDBFieldMax,
			ErrIdDBFieldMin,
			ErrIdDBFieldRange,
			ErrIdDBFieldUnDefined,
			ErrIdDBQueParams,
			ErrIdDBQueNone,
			ErrIdDBQueForeignNone,
		},
	}

	// ErrMsgPatterns 错误模式匹配
	ErrMsgPatterns = map[string]string{
		"duplicate key value violates unique constraint": ErrIdDBPkDuplicated,
	}
)
