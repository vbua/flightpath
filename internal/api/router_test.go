package api

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/vbua/flightpath/internal/config"
)

var (
	errAccept = errors.New("accept error")
	errClose  = errors.New("close error")
)

type routerProvider interface {
	Start(listener net.Listener) error
	Shutdown() error
}

type acceptErrorListener struct{}

func (b acceptErrorListener) Accept() (net.Conn, error) {
	return nil, errAccept
}

func (b acceptErrorListener) Close() error {
	return nil
}

func (b acceptErrorListener) Addr() net.Addr {
	return nil
}

func RunStartErrorTest(t *testing.T, router routerProvider, errorStr string) {
	t.Helper()

	assert.EqualError(t, router.Start(acceptErrorListener{}), errorStr)
}

func TestStartError(t *testing.T) {
	RunStartErrorTest(t, NewMainRouter(
		config.ServerOpts{},
		nil,
	), "main router listen: accept error")
}

func RunShutdownErrorTest(t *testing.T, router routerProvider, errorStr string) {
	t.Helper()

	listener := newCloseErrorListener()

	wgr := &sync.WaitGroup{}
	wgr.Add(1)

	go func() {
		defer wgr.Done()

		assert.NoError(t, router.Start(listener))
	}()

	listener.WaitForAccept()

	err := router.Shutdown()
	assert.EqualError(t, err, errorStr)

	wgr.Wait()
}

type closeErrorListener struct {
	accepted chan struct{}
	listener *fasthttputil.InmemoryListener
}

func newCloseErrorListener() *closeErrorListener {
	return &closeErrorListener{
		listener: fasthttputil.NewInmemoryListener(),
		accepted: make(chan struct{}),
	}
}

func (b *closeErrorListener) Accept() (net.Conn, error) {
	close(b.accepted)

	conn, err := b.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("can't accept: %w", err)
	}

	return conn, nil
}

func (b *closeErrorListener) WaitForAccept() {
	<-b.accepted
}

func (b *closeErrorListener) Close() error {
	if err := b.listener.Close(); err != nil {
		return fmt.Errorf("can't close listener: %w", err)
	}

	return errClose
}

func (b *closeErrorListener) Addr() net.Addr {
	return b.listener.Addr()
}

func TestShutdownError(t *testing.T) {
	RunShutdownErrorTest(t, NewMainRouter(
		config.ServerOpts{},
		nil,
	), "can't shutdown fasthttp server: close error")
}
