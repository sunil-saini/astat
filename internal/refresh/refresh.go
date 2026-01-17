package refresh

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/model"
	"github.com/sunil-saini/astat/internal/registry"
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

	cache.LockMeta()
	cache.Read(metaFile, &meta)
	if meta.Services == nil {
		meta.Services = make(map[string]cache.ServiceMeta)
	}
	sMeta := meta.Services[resource]
	sMeta.Refreshing = true
	sMeta.BusyPID = os.Getpid()
	meta.Services[resource] = sMeta
	if err := cache.Write(metaFile, &meta); err != nil {
		logger.Error("Failed to update cache metadata: %v", err)
	}
	cache.UnlockMeta()

	// Track whether the refresh was successful
	success := false
	defer func() {
		cache.LockMeta()
		defer cache.UnlockMeta()

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
		if err := cache.Write(metaFile, &meta); err != nil {
			logger.Error("Failed to update cache metadata: %v", err)
		}
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

func RefreshSync[T any](ctx context.Context, resource string, fetch func(ctx context.Context, cfg sdkaws.Config) ([]T, error)) {
	multi := pterm.DefaultMultiPrinter
	multi.Start()
	defer multi.Stop()

	RefreshWithMulti(ctx, resource, fetch, &multi)
}

func RefreshWithMulti[T any](ctx context.Context, resource string, fetch func(ctx context.Context, cfg sdkaws.Config) ([]T, error), multi *pterm.MultiPrinter) {
	var meta cache.Meta
	cache.Read(cache.Path(cache.Dir(), "meta"), &meta)

	if sMeta, ok := meta.Services[resource]; ok && sMeta.Refreshing && IsProcessAlive(sMeta.BusyPID) {
		logger.Info("cache refresh for %s is already ongoing in another terminal (PID: %d)", resource, sMeta.BusyPID)
		return
	}

	s := createSpinner(multi.NewWriter(), resource)
	Refresh(ctx, resource, fetch, &ptermTracker{spinner: s})
}

func createSpinner(writer io.Writer, resource string) *pterm.SpinnerPrinter {
	s, _ := pterm.DefaultSpinner.
		WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
		WithStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithWriter(writer).
		WithRemoveWhenDone(false).
		Start(pterm.LightCyan(fmt.Sprintf("%s pending...", resource)))

	s.SuccessPrinter = pterm.Success.WithPrefix(pterm.Prefix{Text: " ✓ ", Style: pterm.NewStyle(pterm.FgLightGreen)}).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightGreen))
	s.FailPrinter = pterm.Error.WithPrefix(pterm.Prefix{Text: " ✗ ", Style: pterm.NewStyle(pterm.FgLightRed)}).
		WithMessageStyle(pterm.NewStyle(pterm.FgLightRed))

	return s
}

func refreshInternal(ctx context.Context, name string, tracker Tracker) {
	var service *registry.Service
	for _, s := range registry.Registry {
		if s.Name == name {
			service = &s
			break
		}
	}

	if service == nil {
		logger.Warn("unknown service: %s", name)
		return
	}

	Refresh(ctx, name, func(ctx context.Context, cfg sdkaws.Config) ([]any, error) {
		res, err := service.Fetch(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return model.ToAnySlice(res), nil
	}, tracker)
}

var bgWG sync.WaitGroup
var bgCount int32

func Wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		bgWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		select {
		case <-done:
			return
		default:
		}

		if atomic.LoadInt32(&bgCount) > 0 {
			pterm.Println()
			logger.Warn("interrupted: background refresh stopped")
		}
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
		refreshInternal(ctx, service, &silentTracker{})
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
		atomic.AddInt32(&bgCount, 1)
		bgWG.Go(func() {
			defer atomic.AddInt32(&bgCount, -1)
			refreshInternal(ctx, service, &silentTracker{})
		})
	}
}

type silentTracker struct{}

func (s *silentTracker) Update(msg string) {
	_ = msg
}
func (s *silentTracker) Success(msg string) {
	_ = msg
}
func (s *silentTracker) Error(msg string) { logger.Error("%s", msg) }

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
