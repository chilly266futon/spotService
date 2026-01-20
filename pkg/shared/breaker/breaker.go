package breaker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	MaxRequests uint32        // Максимальное количество запросов в полуоткрытом состоянии
	Interval    time.Duration // Интервал для сброса счетчиков
	Timeout     time.Duration // Таймаут для перехода в полуоткрытое состояние
	ReadyToTrip func(counts gobreaker.Counts) bool
}

func DefaultConfig() Config {
	return Config{
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	}
}

func UnaryClientInterceptor(cfg Config) grpc.UnaryClientInterceptor {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "grpc-client",
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: cfg.ReadyToTrip,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// TODO: интеграция с внешним мониторингом
		},
	})

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		_, err := cb.Execute(func() (any, error) {
			err := invoker(ctx, method, req, reply, cc, opts...)

			// Считываем ошибки, которые должны триггерить breaker
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unavailable, codes.Internal, codes.DeadlineExceeded:
						return nil, err
					}
				}
			}

			return nil, err
		})

		return err
	}
}

type Wrapper struct {
	cb *gobreaker.CircuitBreaker
}

func NewWrapper(name string, cfg Config) *Wrapper {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          name,
		MaxRequests:   cfg.MaxRequests,
		Interval:      cfg.Interval,
		Timeout:       cfg.Timeout,
		ReadyToTrip:   cfg.ReadyToTrip,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {},
	})

	return &Wrapper{cb: cb}
}

func (w *Wrapper) Execute(fn func() error) error {
	_, err := w.cb.Execute(func() (any, error) {
		return nil, fn()
	})
	return err
}

func (w *Wrapper) State() gobreaker.State {
	return w.cb.State()
}
