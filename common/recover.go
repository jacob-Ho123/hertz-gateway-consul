package common

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.opentelemetry.io/otel/trace"
)

func RecoveryHandler(ctx context.Context, c *app.RequestContext, err interface{}, stack []byte) {
	traceID := trace.SpanContextFromContext(ctx).TraceID().String()
	hlog.SystemLogger().CtxErrorf(ctx, "[Recovery] traceID=%s err=%v\nstack=%s", traceID, err, stack)
	hlog.SystemLogger().Infof("Client: %s", c.Request.Header.UserAgent())
	//go sendImAlertMessage(traceID, string(stack), string(c.Method()), c.URI().String(), c.Request.Body(), err)
	c.AbortWithStatus(consts.StatusInternalServerError)
}
