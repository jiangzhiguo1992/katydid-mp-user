package middleware

const (
	RequestIDKey     = "RequestID"    // 本地获取的请求ID的上下文键
	XRequestIDHeader = "X-Request-ID" // 网关生成并塞入的requestID
)
