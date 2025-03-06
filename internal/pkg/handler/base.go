package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/i18n"
	"katydid-mp-user/pkg/valid"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Base 处理器基类
type Base struct {
	gCtx *gin.Context
	Conf
	*DB
	*Auth
	*App
}

// Conf 配置
type Conf struct {
	Lang string // 语言
}

// DB 数据库
type DB struct {
	InCache *sync.Map // 内存缓存 TODO:GG 找个库?
	DBCache *gorm.DB  // 数据库缓存
	DBRepo  *gorm.DB  // 数据库仓储
	DBTx    *gorm.Tx  // 数据库事务
}

// Auth 身份验证
type Auth struct {
	OrgID  *uint64 // 组织ID
	RoleID *string // 用户角色
	AccID  *uint64 // 账号ID
	UserID *uint64 // 用户ID
}

// App 应用
type App struct {
	AppID *uint64 // 应用Id
	CltID *uint64 // 客户端Id
	VerID *uint64 // 版本Id
}

// NewBase 创建新的基础处理器
func NewBase(db *DB) *Base {
	return &Base{
		gCtx: nil,
		Conf: Conf{
			Lang: i18n.DefLang(),
		},
		DB:   db,
		Auth: nil,
		App:  nil,
	}
}

// SetGCtx 设置原始gin上下文
func (b *Base) SetGCtx(gCtx *gin.Context) {
	b.gCtx = gCtx
	if b.gCtx == nil {
		b.Lang = i18n.DefLang()
		return
	}
	b.Lang = b.gCtx.GetHeader("Use-Language")
}

// GCtx 获取原始gin上下文
func (b *Base) GCtx() *gin.Context {
	return b.gCtx
}

/********************************************************************************
 *********************************** Request ************************************
 ********************************************************************************/

// RequestBind 绑定并验证请求数据
func (b *Base) RequestBind(obj any, must bool) *err.CodeErrs {
	// bind会自动推断type
	if must {
		if e := b.gCtx.Bind(obj); e != nil {
			return err.Match(e).WrapLocalize("invalid_request_format", nil, nil).Real()
		}
	} else {
		if e := b.gCtx.ShouldBind(obj); e != nil {
			return err.Match(e).WrapLocalize("invalid_request_format", nil, nil).Real()
		}
	}
	return valid.Check(obj, valid.SceneBind)
}

// RequestParam 获取路径参数
func (b *Base) RequestParam(key string, defVal string) string {
	value := b.gCtx.Param(key)
	if value == "" {
		return defVal
	}
	return value
}

// RequestQuery 获取查询参数
func (b *Base) RequestQuery(key string, defVal string) (string, bool) {
	value, exists := b.gCtx.GetQuery(key)
	if !exists {
		return defVal, exists
	}
	return value, exists
}

// RequestPagination 获取分页参数
func (b *Base) RequestPagination() (page, size int) {
	pageStr, _ := b.RequestQuery("page", "1")
	sizeStr, _ := b.RequestQuery("pageSize", "20")

	page, _ = strconv.Atoi(pageStr)
	size, _ = strconv.Atoi(sizeStr)

	if page < 1 {
		page = 1
	}
	if size < 1 {
		page = 20
	}
	if size > 100 {
		page = 100
	}
	return page, size
}

// RequestSorting 获取排序参数
func (b *Base) RequestSorting(defField, defOrder string, fieldRanges []string) (field, order string) {
	field, _ = b.RequestQuery("sortBy", defField)
	order, _ = b.RequestQuery("sortOrder", defOrder)

	find := false
	for _, v := range fieldRanges {
		if field == v {
			find = true
			break
		}
	}
	if !find {
		field = defField
	}

	if order != "asc" && order != "desc" {
		order = defOrder
	}
	return field, order
}

/********************************************************************************
 *********************************** Response ***********************************
 ********************************************************************************/

// Response200 成功响应
func (b *Base) Response200(data any) {
	b.Response(http.StatusOK, 0, "success", data)
}

// Response201 创建成功响应
func (b *Base) Response201(data any) {
	b.Response(http.StatusCreated, 0, "created_success", data)
}

// Response400 请求错误响应
func (b *Base) Response400(msg string, data any) {
	if msg == "" {
		msg = "bad_request"
	}
	localizedMsg := i18n.LocalizeTry(b.Lang, msg, nil)
	b.Response(http.StatusBadRequest, 0, localizedMsg, data)
}

// Response401 未授权响应
func (b *Base) Response401(data any) {
	msg := i18n.LocalizeTry(b.Lang, "unauthorized", nil)
	b.Response(http.StatusUnauthorized, 0, msg, data)
}

// Response403 禁止访问响应
func (b *Base) Response403(msg string) {
	if msg == "" {
		msg = "forbidden"
	}
	localizedMsg := i18n.LocalizeTry(b.Lang, msg, nil)
	b.Response(http.StatusForbidden, 403, localizedMsg, nil)
}

// Response404 资源不存在响应
func (b *Base) Response404(msg string) {
	if msg == "" {
		msg = "not_found"
	}
	localizedMsg := i18n.LocalizeTry(b.Lang, msg, nil)
	b.Response(http.StatusNotFound, 404, localizedMsg, nil)
}

func (b *Base) Response(status, code int, msg string, data any) {
	if e, ok := data.(error); ok {
		b.responseErr(status, code, e)
		return
	}
	b.responseData(status, code, msg, data)
}

func (b *Base) responseErr(status, code int, data error) {
	var cErr *err.CodeErrs
	var v *err.CodeErrs
	if errors.As(data, &v) {
		cErr = v
	} else {
		cErr = err.Match(data).Real()
	}
	if code == 0 {
		code = cErr.Code()
	}

	msg := cErr.ToLocales(func(localize string, template1s []any, template2s map[string]any) string {
		var templates []any
		for _, v := range template1s {
			if _, ok := v.(string); !ok {
				templates = append(templates, v)
				continue
			}
			temp := i18n.LocalizeTry(b.Lang, v.(string), nil)
			templates = append(templates, temp)
		}
		r1 := i18n.LocalizeTry(b.Lang, localize, template2s)
		return fmt.Sprintf(r1, templates...)
	})

	if len(msg) == 0 {
		msg = i18n.LocalizeTry(b.Lang, "unknown_err", nil)
	}

	b.responseData(status, code, msg, nil)
}

func (b *Base) responseData(status, code int, msg string, data any) {
	// TODO:GG 有些字段，返回的时候是要忽略的(利用json:"-"来做吗?)
	body := gin.H{"code": code, "msg": msg, "data": data}

	accept := b.gCtx.GetHeader("Accept")
	if accept == "" || strings.Contains(accept, "*/*") || strings.Contains(accept, "application/*") {
		accept = binding.MIMEJSON
	} else if strings.Contains(accept, "msg/*") {
		accept = binding.MIMEXML
	}

	switch {
	case strings.Contains(accept, binding.MIMEJSON):
		b.gCtx.JSON(status, body)
	case strings.Contains(accept, binding.MIMEPROTOBUF):
		b.gCtx.ProtoBuf(status, body)
	case strings.Contains(accept, binding.MIMEHTML):
		b.gCtx.HTML(status, "", msg)
	case strings.Contains(accept, binding.MIMEXML), strings.Contains(accept, binding.MIMEXML2):
		b.gCtx.XML(status, body)
	case strings.Contains(accept, binding.MIMETOML):
		b.gCtx.TOML(status, body)
	case strings.Contains(accept, binding.MIMEYAML), strings.Contains(accept, binding.MIMEYAML2):
		b.gCtx.YAML(status, body)
	default:
		b.gCtx.String(status, msg)
	}
}

//// HasPermission 检查是否有指定权限
//func (b *Base) HasPermission(perm string) bool {
//	for _, p := range b.Perms {
//		if p == perm || p == "*" {
//			return true
//		}
//	}
//	return false
//}
//
//// RequirePermission 要求特定权限, 无权限时返回错误
//func (b *Base) RequirePermission(perm string) bool {
//	if !b.HasPermission(perm) {
//		b.Response401("")
//		return false
//	}
//	return true
//}
//
//// RequireAuthenticated 要求已认证用户
//func (b *Base) RequireAuthenticated() bool {
//	if b.AccID == 0 {
//		b.Response401("login_required")
//		return false
//	}
//	return true
//}
//
//// IsAdmin 判断是否是管理员
//func (b *Base) IsAdmin() bool {
//	return *b.Role == "admin" || b.HasPermission("admin")
//}
//
//// RequireAdmin 要求管理员权限
//func (b *Base) RequireAdmin() bool {
//	if !b.IsAdmin() {
//		b.Response403("admin_required")
//		return false
//	}
//	return true
//}
//
//// IsCurrentOrg 检查是否与指定组织匹配
//func (b *Base) IsCurrentOrg(orgId uint64) bool {
//	return *b.OrgId == orgId
//}
//
//// RequireOrg 要求属于指定组织
//func (b *Base) RequireOrg(orgId uint64) bool {
//	if !b.IsCurrentOrg(orgId) && !b.IsAdmin() {
//		b.Response403("org_access_denied")
//		return false
//	}
//	return true
//}
