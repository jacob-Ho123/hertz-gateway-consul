package middleware

import (
	"context"
	"github.com/hertz-contrib/obs-opentelemetry/tracing"
	"github.com/jacob-Ho123/hertz-gateway-consul/common"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	hertzConfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hertz-contrib/registry/consul"
	"github.com/hertz-contrib/reverseproxy"
)

func ReverseProxyMiddleware(config *common.ProxyConfig) app.HandlerFunc {

	// 根据环境获取consul配置
	consulAddr, consulDC := config.Common.ConsulAddr, config.Common.ConsulDc

	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulAddr
	if consulDC != "" {
		consulConfig.Datacenter = consulDC
	}

	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		hlog.Errorf("NewHertzServer err: %v", err)
		panic(err)
	}

	// build a consul resolver with the consul client
	resolver := consul.NewConsulResolver(consulClient)

	// build a hertz client with the consul resolver
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(tracing.ClientMiddleware())
	cli.Use(sd.Discovery(resolver))

	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())

		// 获取路径的第一段作为服务名
		serviceName := getFirstPathSegment(path)
		hlog.Infof("get serviceName: %s", serviceName)

		if serviceName == "" {
			c.JSON(400, map[string]interface{}{
				"error": "invalid service path",
			})
			c.Abort()
			return
		}
		proxy, _err := reverseproxy.NewSingleHostReverseProxy(serviceName)
		if _err != nil {
			c.JSON(500, map[string]interface{}{
				"error": _err.Error(),
			})
			c.Abort()
			return
		}
		proxy.SetClient(cli)
		proxy.SetDirector(func(req *protocol.Request) {
			hlog.Debugf("request host: %s,path: %s", req.Host(), req.URI().Path())
			hlog.Debugf("request path: %s", reverseproxy.JoinURLPath(req, proxy.Target))
			req.SetHost(serviceName)
			req.Options().Apply([]hertzConfig.RequestOption{hertzConfig.WithSD(true)})
			hlog.Debugf("request host: %s,path: %s", req.Host(), req.URI().Path())
		})
		proxy.ServeHTTP(ctx, c)
		c.Abort()

	}
}

// getFirstPathSegment 获取路径的第一段作为服务名
// 例如: /user-service/api/v1/users -> user-service
func getFirstPathSegment(path string) string {
	// 移除开头的斜杠
	path = strings.TrimPrefix(path, "/")

	// 如果路径为空，返回空字符串
	if path == "" {
		return ""
	}

	// 按斜杠分割路径
	segments := strings.SplitN(path, "/", 2)

	// 返回第一段
	return segments[0]
}
