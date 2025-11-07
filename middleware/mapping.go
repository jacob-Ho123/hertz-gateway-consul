package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/jacob-Ho123/hertz-gateway-consul/common"
	"strings"
)

func MappingMiddleware(config *common.ProxyConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		appId := strings.TrimSpace(string(c.GetHeader("app")))
		if appId == "" {
			c.Next(ctx)
			return
		}
		reqReplacer, hasReqReplacer := config.ReqReplacerMap[appId]
		respReplacer, hasRespReplacer := config.RespReplacerMap[appId]
		if hasReqReplacer || hasRespReplacer {
			doReplace(ctx, c, reqReplacer, respReplacer)
		} else {
			c.Next(ctx)
		}
	}
}

func doReplace(ctx context.Context, c *app.RequestContext, requestReplacer, responseReplacer *strings.Replacer) {
	// 只处理 POST 请求和 JSON 内容
	if !c.Request.Header.IsPost() ||
		!strings.Contains(string(c.Request.Header.ContentType()), "application/json") {
		c.Next(ctx)
		return
	}

	// 读取原始请求体
	if requestReplacer != nil {
		body := c.GetRequest().Body()
		if len(body) > 0 {
			// 压缩 JSON
			var compacted bytes.Buffer
			if err := json.Compact(&compacted, body); err == nil {
				// 替换 JSON keys
				replaced := requestReplacer.Replace(compacted.String())
				// 更新请求体
				c.Request.SetBodyString(replaced)
			} else {
				hlog.Errorf("Failed to compress request body: %v", err)
			}
		}
	}

	// 继续处理请求
	c.Next(ctx)

	if responseReplacer != nil {
		respBody := c.GetResponse().Body()
		if len(respBody) >= 0 {
			// 压缩 JSON
			var respCompacted bytes.Buffer
			if err := json.Compact(&respCompacted, respBody); err == nil {
				// 替换 JSON keys
				respReplaced := responseReplacer.Replace(respCompacted.String())
				c.SetBodyString(respReplaced)
			} else {
				hlog.Errorf("Failed to compress request body: %v", err)
			}
		}
	}

}
