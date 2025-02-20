package error

const (
	CodeUnknown = 0
	CodeDB      = 1000

	MsgIdDBPkDuplicated   = "err_db_pk_duplicated"
	MsgIdDBAddNil         = "err_db_add_nil"
	MsgIdDBDelNil         = "err_db_del_nil"
	MsgIdDBUpdNil         = "err_db_upd_nil"
	MsgIdDBQueNil         = "err_db_que_nil"
	MsgIdDBFieldNil       = "err_db_field_nil"
	MsgIdDBFieldLarge     = "err_db_field_large"
	MsgIdDBFieldShort     = "err_db_field_short"
	MsgIdDBFieldMax       = "err_db_field_max"
	MsgIdDBFieldMin       = "err_db_field_min"
	MsgIdDBFieldRange     = "err_db_field_range"
	MsgIdDBFieldUnDefined = "err_db_field_undefined"
	MsgIdDBQueParams      = "err_db_que_params"
	MsgIdDBQueNone        = "err_db_que_none"
	MsgIdDBQueForeignNone = "err_db_que_foreign_none"
)

var (
	// 错误信息映射
	codeMsgIds = map[int][]string{
		CodeDB: {
			MsgIdDBPkDuplicated,
			MsgIdDBAddNil,
			MsgIdDBDelNil,
			MsgIdDBUpdNil,
			MsgIdDBQueNil,
			MsgIdDBFieldNil,
			MsgIdDBFieldLarge,
			MsgIdDBFieldShort,
			MsgIdDBFieldMax,
			MsgIdDBFieldMin,
			MsgIdDBFieldRange,
			MsgIdDBFieldUnDefined,
			MsgIdDBQueParams,
			MsgIdDBQueNone,
			MsgIdDBQueForeignNone,
		},
	}

	// 错误模式匹配
	errorPatterns = map[string]string{
		"duplicate key value violates unique constraint": MsgIdDBPkDuplicated,
	}
)
