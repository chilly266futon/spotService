package interceptors

import (
	"context"
	"google.golang.org/grpc/metadata"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		traceID := GetTraceID(ctx)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				zap.String("grpc_code", st.Code().String()),
				zap.String("error", err.Error()),
			)
			logger.Error("grpc request failed", fields...)
		} else {
			logger.Info("grpc request completed", fields...)
		}

		return resp, err
	}
}

func GetTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	if values := md.Get("x-request-id"); len(values) > 0 {
		return values[0]
	}

	return ""
}
