package refresh

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/model"
)

type Tracker interface {
	Update(msg string)
	Success(msg string)
	Error(msg string)
}

type ptermTracker struct {
	spinner *pterm.SpinnerPrinter
}

func (p *ptermTracker) Update(msg string)  { p.spinner.UpdateText(msg) }
func (p *ptermTracker) Success(msg string) { p.spinner.Success(msg) }
func (p *ptermTracker) Error(msg string)   { p.spinner.Fail(msg) }

func Refresh[T any](ctx context.Context, resource string, fetch func(ctx context.Context, cfg sdkaws.Config) ([]T, error), tracker Tracker) {
	tracker.Update(fmt.Sprintf("%s loading config...", resource))
	cfg, err := aws.LoadConfig(ctx)
	if err != nil {
		tracker.Error(fmt.Sprintf("%s config failed: %v", resource, err))
		return
	}

	dir := cache.Dir()
	cache.EnsureDir(dir)

	metaFile := cache.Path(dir, "meta")
	var meta cache.Meta
	cache.Read(metaFile, &meta)
	if meta.Services == nil {
		meta.Services = make(map[string]cache.ServiceMeta)
	}
	sMeta := meta.Services[resource]
	sMeta.Refreshing = true
	sMeta.BusyPID = os.Getpid()
	meta.Services[resource] = sMeta
	cache.Write(metaFile, meta)

	// Track whether the refresh was successful
	success := false
	defer func() {
		cache.Read(metaFile, &meta)
		if meta.Services == nil {
			meta.Services = make(map[string]cache.ServiceMeta)
		}
		sMeta := meta.Services[resource]
		sMeta.Refreshing = false
		// Only update LastUpdated if the refresh was successful
		if success {
			sMeta.LastUpdated = time.Now()
			meta.LastUpdated = time.Now()
		}
		meta.Services[resource] = sMeta
		cache.Write(metaFile, meta)
	}()

	tracker.Update(fmt.Sprintf("%s fetching...", resource))
	data, err := fetch(ctx, cfg)
	if err != nil {
		tracker.Error(fmt.Sprintf("%s fetch failed: %v", resource, err))
		return
	}

	tracker.Update(fmt.Sprintf("%s saving...", resource))
	cache.Write(cache.Path(dir, resource), data)
	success = true
	tracker.Success(fmt.Sprintf("%s refreshed", resource))
}

var serviceRegistry = map[string]func(context.Context, sdkaws.Config) (any, error){
	"ec2": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchEC2Instances(ctx, cfg)
	},
	"s3": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchS3Buckets(ctx, cfg)
	},
	"lambda": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchLambdaFunctions(ctx, cfg)
	},
	"cloudfront": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchCloudFront(ctx, cfg)
	},
	"route53-zones": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchHostedZones(ctx, cfg)
	},
	"route53-records": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchAllRoute53Records(ctx, cfg)
	},
	"ssm": func(ctx context.Context, cfg sdkaws.Config) (any, error) {
		return aws.FetchSSMParameters(ctx, cfg)
	},
}

func RefreshSync[T any](ctx context.Context, resource string, fetch func(ctx context.Context, cfg sdkaws.Config) ([]T, error)) {
	var meta cache.Meta
	cache.Read(cache.Path(cache.Dir(), "meta"), &meta)

	if sMeta, ok := meta.Services[resource]; ok && sMeta.Refreshing && IsProcessAlive(sMeta.BusyPID) {
		logger.Info("cache refresh for %s is already ongoing in another terminal (PID: %d)", resource, sMeta.BusyPID)
		return
	}

	multi := pterm.DefaultMultiPrinter
	multi.Start()
	defer multi.Stop()

	s, _ := pterm.DefaultSpinner.
		WithSequence("⠃", "⠉", "⠘", "⠒").
		WithStyle(pterm.NewStyle(pterm.FgCyan)).
		WithMessageStyle(pterm.NewStyle(pterm.FgCyan)).
		WithWriter(multi.NewWriter()).
		Start(fmt.Sprintf("%s pending...", resource))

	s.SuccessPrinter = pterm.Success.WithPrefix(pterm.Prefix{Text: " ✓ ", Style: pterm.NewStyle(pterm.FgGreen)})
	s.FailPrinter = pterm.Error.WithPrefix(pterm.Prefix{Text: " ✗ ", Style: pterm.NewStyle(pterm.FgRed)})

	Refresh(ctx, resource, fetch, &ptermTracker{spinner: s})
}

// RefreshWithMulti refreshes a service using a shared multi-printer for concurrent execution
func RefreshWithMulti[T any](ctx context.Context, resource string, fetch func(ctx context.Context, cfg sdkaws.Config) ([]T, error), multi *pterm.MultiPrinter) {
	var meta cache.Meta
	cache.Read(cache.Path(cache.Dir(), "meta"), &meta)

	if sMeta, ok := meta.Services[resource]; ok && sMeta.Refreshing && IsProcessAlive(sMeta.BusyPID) {
		logger.Info("cache refresh for %s is already ongoing in another terminal (PID: %d)", resource, sMeta.BusyPID)
		return
	}

	s, _ := pterm.DefaultSpinner.
		WithSequence("⠃", "⠉", "⠘", "⠒").
		WithStyle(pterm.NewStyle(pterm.FgCyan)).
		WithMessageStyle(pterm.NewStyle(pterm.FgCyan)).
		WithWriter(multi.NewWriter()).
		Start(fmt.Sprintf("%s pending...", resource))

	s.SuccessPrinter = pterm.Success.WithPrefix(pterm.Prefix{Text: " ✓ ", Style: pterm.NewStyle(pterm.FgGreen)})
	s.FailPrinter = pterm.Error.WithPrefix(pterm.Prefix{Text: " ✗ ", Style: pterm.NewStyle(pterm.FgRed)})

	Refresh(ctx, resource, fetch, &ptermTracker{spinner: s})
}

func refreshInternal(ctx context.Context, name string, tracker Tracker) {
	fetch, ok := serviceRegistry[name]
	if !ok {
		logger.Warn("unknown service: %s", name)
		return
	}

	Refresh(ctx, name, func(ctx context.Context, cfg sdkaws.Config) ([]any, error) {
		res, err := fetch(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return model.ToAnySlice(res), nil
	}, tracker)
}

var bgWG sync.WaitGroup

func Wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		bgWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		logger.Warn("interrupting background refresh wait")
	}
}

func AutoRefreshIfStale(ctx context.Context, service string) {
	if service == "" {
		return
	}

	var meta cache.Meta
	metaPath := cache.Path(cache.Dir(), "meta")
	err := cache.Read(metaPath, &meta)
	if err != nil {
		logger.Warn("cache not initialized, triggering background refresh")
		bgWG.Add(1)
		go func() {
			defer bgWG.Done()
			refreshInternal(context.Background(), service, &silentTracker{})
		}()
		return
	}

	if sMeta, ok := meta.Services[service]; ok && sMeta.Refreshing && IsProcessAlive(sMeta.BusyPID) {
		logger.Info("background refresh for %s is already ongoing (PID: %d)", service, sMeta.BusyPID)
		return
	}

	ttl := viper.GetDuration("ttl")
	sMeta, ok := meta.Services[service]
	if !ok || time.Since(sMeta.LastUpdated) > ttl {
		logger.Info("service %s is stale, auto-refreshing...", service)
		bgWG.Add(1)
		go func() {
			defer bgWG.Done()
			refreshInternal(context.Background(), service, &silentTracker{})
		}()
	}
}

type silentTracker struct{}

func (s *silentTracker) Update(msg string)  {}
func (s *silentTracker) Success(msg string) {}
func (s *silentTracker) Error(msg string)   { logger.Error("%s", msg) }

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
