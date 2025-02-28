package httpc

import "sync"

var (
	// 默认对象
	defaultHttpReq *VClient
	// 对象池
	httpReqClt = &sync.Pool{
		New: func() interface{} {
			return new(VClient)
		},
	}
)
