package core

import (
	"context"
	"sync"
	"time"

	"github.com/aler9/rtsp-simple-server/internal/logger"
	"github.com/aler9/rtsp-simple-server/internal/rpicamera"
)

const (
	rpiCameraSourceRetryPause = 5 * time.Second
)

type rpiCameraSourceParent interface {
	log(logger.Level, string, ...interface{})
	onSourceStaticSetReady(req pathSourceStaticSetReadyReq) pathSourceStaticSetReadyRes
	onSourceStaticSetNotReady(req pathSourceStaticSetNotReadyReq)
}

type rpiCameraSource struct {
	wg     *sync.WaitGroup
	parent rpiCameraSourceParent

	ctx       context.Context
	ctxCancel func()
}

func newRPICameraSource(
	parentCtx context.Context,
	wg *sync.WaitGroup,
	parent rpiCameraSourceParent,
) *rpiCameraSource {
	ctx, ctxCancel := context.WithCancel(parentCtx)

	s := &rpiCameraSource{
		wg:        wg,
		parent:    parent,
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}

	s.Log(logger.Info, "started")

	s.wg.Add(1)
	go s.run()

	return s
}

func (s *rpiCameraSource) close() {
	s.Log(logger.Info, "stopped")
	s.ctxCancel()
}

func (s *rpiCameraSource) Log(level logger.Level, format string, args ...interface{}) {
	s.parent.log(level, "[rpicamera source] "+format, args...)
}

func (s *rpiCameraSource) run() {
	defer s.wg.Done()

outer:
	for {
		innerCtx, innerCtxCancel := context.WithCancel(context.Background())
		innerErr := make(chan error)
		go func() {
			innerErr <- s.runInner(innerCtx)
		}()

		select {
		case err := <-innerErr:
			innerCtxCancel()
			s.Log(logger.Info, "ERR: %v", err)

		case <-s.ctx.Done():
			innerCtxCancel()
			<-innerErr
		}

		select {
		case <-time.After(rpiCameraSourceRetryPause):
		case <-s.ctx.Done():
			break outer
		}
	}

	s.ctxCancel()
}

func (s *rpiCameraSource) runInner(innerCtx context.Context) error {
	_, err := rpicamera.New()
	return err
}

// onSourceAPIDescribe implements source.
func (*rpiCameraSource) onSourceAPIDescribe() interface{} {
	return struct {
		Type string `json:"type"`
	}{"rpiCameraSource"}
}
