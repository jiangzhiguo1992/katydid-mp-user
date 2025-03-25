package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"net/http"
	"net/url"
	"strings"
)

const (
	csrfCookieName = "csrf_token"   // CSRF令牌的Cookie名称
	csrfHeaderName = "X-CSRF-Token" // CSRF令牌的Header名称
	csrfFieldName  = "_csrf"        // CSRF令牌的表单字段名称
)

var (
	// 用于加密CSRF令牌的密钥
	csrfSecureCookie *securecookie.SecureCookie
)

// CSRFOptions 配置CSRF中间件的选项
type CSRFOptions struct {
	CookieMaxAge   int           // Cookie有效期(秒)
	Path           string        // 路径
	Domain         string        // 域名
	Secure         bool          // 是否仅HTTPS
	HttpOnly       bool          // 是否HttpOnly
	SameSite       http.SameSite // SameSite策略
	SetTokenHeader bool          // 是否在响应头中设置令牌
	ExcludePaths   []string      // 要排除的路径前缀
	CheckReferer   bool          // 是否检查Referer头
	AllowedHosts   []string      // 允许的主机名列表
}

// DefaultCSRFOptions 返回默认CSRF选项
func DefaultCSRFOptions(excludes []string) *CSRFOptions {
	return &CSRFOptions{
		CookieMaxAge:   3600, // 1小时
		Path:           "/",
		Domain:         "",
		Secure:         true,
		HttpOnly:       true,
		SameSite:       http.SameSiteStrictMode, // 限制跨站请求携带Cookie
		SetTokenHeader: true,
		ExcludePaths:   excludes, // []string{"/api/"},
		CheckReferer:   true,     // 默认开启Referer检查
		AllowedHosts:   []string{},
	}
}

// CSRF 跨站请求伪造
func CSRF(hashKey, blockKey []byte, options ...*CSRFOptions) gin.HandlerFunc {
	// 初始化CSRF安全Cookie
	if hashKey != nil && blockKey != nil {
		csrfSecureCookie = securecookie.New(hashKey, blockKey)
	} else {
		// 如果没有提供密钥，生成随机密钥(仅用于开发环境)
		hashKey = securecookie.GenerateRandomKey(32)
		blockKey = securecookie.GenerateRandomKey(32)
		if hashKey != nil && blockKey != nil {
			csrfSecureCookie = securecookie.New(hashKey, blockKey)
		}
	}

	// 使用默认配置或用户提供的配置
	var opts *CSRFOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = DefaultCSRFOptions([]string{"/api/"})
	}

	return func(c *gin.Context) {
		// 检查是否应该排除此路径
		path := c.Request.URL.Path
		for _, prefix := range opts.ExcludePaths {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		// 只处理非安全方法(POST, PUT, DELETE, PATCH)的CSRF保护
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			// 对于安全方法，设置CSRF令牌
			setCSRFToken(c, opts)
			c.Next()
			return
		}

		// 对于非安全方法，先检查Referer
		if opts.CheckReferer && !isValidReferer(c.Request, opts.AllowedHosts) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "无效的请求来源",
			})
			return
		}

		// 对于非安全方法，验证CSRF令牌
		if !validateCSRFToken(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF令牌验证失败",
			})
			return
		}

		// 验证通过后，重新设置令牌以延长有效期
		setCSRFToken(c, opts)
		c.Next()
	}
}

// 检查Referer是否有效
func isValidReferer(r *http.Request, allowedHosts []string) bool {
	// 获取Referer
	referer := r.Header.Get("Referer")
	if referer == "" {
		// 没有Referer头，根据安全考虑，默认拒绝
		return false
	}

	// 解析Referer URL
	refererURL, err := url.Parse(referer)
	if err != nil {
		return false
	}

	// 获取当前请求的主机
	requestHost := r.Host

	// 检查Referer的主机是否匹配当前主机
	refererHost := refererURL.Host
	if refererHost == requestHost {
		return true
	}

	// 检查是否在允许的主机列表中
	for _, host := range allowedHosts {
		if refererHost == host {
			return true
		}
	}

	return false
}

// 生成CSRF令牌
func generateCSRFToken() (string, error) {
	// 生成32字节随机数
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// ���码为Base64字符串
	token := base64.StdEncoding.EncodeToString(b)

	// 修剪末尾的等号，使令牌更加URL友好
	token = strings.TrimRight(token, "=")

	return token, nil
}

// 设置CSRF令���
func setCSRFToken(c *gin.Context, opts *CSRFOptions) {
	// 检查是否已有令牌
	_, err := c.Cookie(csrfCookieName)
	if err == nil {
		// 已有令牌，尝试获取并存入上下文
		cookie, _ := c.Cookie(csrfCookieName)
		var storedToken string
		if err := csrfSecureCookie.Decode(csrfCookieName, cookie, &storedToken); err == nil {
			c.Set(CSRFKeyToken, storedToken)
			if opts.SetTokenHeader {
				c.Header(csrfHeaderName, storedToken)
			}
		}
		return
	}

	// 生成新令牌
	token, err := generateCSRFToken()
	if err != nil {
		return
	}

	// 将令牌存储在安全Cookie中
	encoded, err := csrfSecureCookie.Encode(csrfCookieName, token)
	if err != nil {
		return
	}

	// 设置Cookie
	c.SetCookie(
		csrfCookieName,
		encoded,
		opts.CookieMaxAge,
		opts.Path,
		opts.Domain,
		opts.Secure,
		opts.HttpOnly,
	)
	c.SetSameSite(opts.SameSite)

	// 设置响应头，方便前端JS获取
	if opts.SetTokenHeader {
		c.Header(csrfHeaderName, token)
	}

	// 将令牌存入上下文，方便模板渲染
	c.Set(CSRFKeyToken, token)
}

// 验证CSRF令牌
func validateCSRFToken(c *gin.Context) bool {
	// 从Cookie获取存储的令牌
	cookie, err := c.Cookie(csrfCookieName)
	if err != nil {
		return false
	}

	// 解码令牌
	var storedToken string
	if err = csrfSecureCookie.Decode(csrfCookieName, cookie, &storedToken); err != nil {
		return false
	}

	// 从请求中获取令牌(先从Header,再从表单)
	var requestToken string

	// 从Header获取
	requestToken = c.GetHeader(csrfHeaderName)

	// 如果Header中没有，尝试从表单获取
	if requestToken == "" {
		requestToken = c.PostForm(csrfFieldName)
	}

	// 进行比较
	return requestToken != "" && requestToken == storedToken
}
