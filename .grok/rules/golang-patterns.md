# Go Patterns

## Functional Options

```go
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func NewServer(opts ...Option) *Server {
    s := &Server{port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

## Small Interfaces

Define interfaces where they are used, not where they are implemented.

## Dependency Injection

Use constructor functions to inject dependencies:

```go
func NewUserService(repo UserRepository, logger Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}
```

## Outbound HTTP — context timeouts

Every outbound HTTP call passes a `context.Context` with a timeout. No
bare `http.Get`. No implicit timeouts relying on the default client.

```go
ctx, cancel := context.WithTimeout(parent, 10*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

The timeout value is declared once per package as a named constant, not
inlined at the call site. The registry package is the only place that
makes outbound HTTP in v1.

## Concurrency — errgroup + semaphore

All concurrent work uses a bounded worker shape. Never spawn unbounded
goroutines per input.

```go
g, ctx := errgroup.WithContext(ctx)
sem := make(chan struct{}, 8)
for i := range deps {
    i := i
    g.Go(func() error {
        sem <- struct{}{}
        defer func() { <-sem }()
        return fetchOne(ctx, &deps[i])
    })
}
return g.Wait()
```

Bounded at 8 concurrent calls per source. Not global.

`i := i` on every loop that closes over `i` in a goroutine.

Acquire the semaphore inside the goroutine, not before `g.Go`, so
the goroutine count is bounded by the errgroup rather than the
channel.

`pkg/registry` is the only package with concurrent fetches in v1.

## Reference

See skill: `golang-patterns` for comprehensive Go patterns including concurrency, error handling, and package organization.
