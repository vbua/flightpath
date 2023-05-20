package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/vbua/flightpath/internal/config"
	"github.com/vbua/flightpath/internal/models"
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
	RunStartErrorTest(t, NewRouter(
		nil,
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
	RunShutdownErrorTest(t, NewRouter(
		nil,
		config.ServerOpts{},
		nil,
	), "can't shutdown fasthttp server: close error")
}

type RouterSuite struct {
	suite.Suite

	Client  nethttp.Client
	router  *Router
	service *MockFlightService

	wg sync.WaitGroup

	testError error
}

func TestRouter(t *testing.T) {
	suite.Run(t, &RouterSuite{})
}

func (s *RouterSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())

	s.service = NewMockFlightService(ctrl)
	s.router.srv = s.service
}

func (s *RouterSuite) SetupSuite() {
	s.router = NewRouter(
		nil,
		config.ServerOpts{},
		nil,
	)

	listener := fasthttputil.NewInmemoryListener()
	s.Client = nethttp.Client{
		Transport: &nethttp.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				conn, err := listener.Dial()
				if err != nil {
					return nil, fmt.Errorf("can't dial: %w", err)
				}

				return conn, nil
			},
		},
	}

	s.wg = sync.WaitGroup{}

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		err := s.router.Start(listener)
		s.NoError(err)
	}()
}

func (s *RouterSuite) TearDownSuite() {
	err := s.router.Shutdown()
	s.NoError(err)

	s.wg.Wait()
}

func (s *RouterSuite) MakeURL(path string) string {
	host, err := os.Hostname()
	s.NoError(err)

	return fmt.Sprintf("http://%s/%s", host, path)
}

func (s *RouterSuite) PostJSON(path, body string) string {
	URL, err := url.Parse(s.MakeURL(path))
	s.NoError(err)

	request := &nethttp.Request{
		Method: nethttp.MethodPost,
		URL:    URL,
		Header: nethttp.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer([]byte(body))),
	}

	resp, err := s.Client.Do(request)
	s.NoError(err)

	respBody, err := io.ReadAll(resp.Body)
	s.NoError(err)

	s.NoError(resp.Body.Close())

	return string(respBody)
}

func (s *RouterSuite) TestCalculate_Ok() {
	s.service.EXPECT().FindStartAndEndOfPath([][]string{{"SFO", "EWR"}}).Return(models.Flight{
		Source:      "SFO",
		Destination: "EWR",
	})

	resp := s.PostJSON("calculate", `{"flights":[["SFO", "EWR"]]}`)
	s.Equal(`{"source":"SFO","destination":"EWR"}`, resp)
}

func (s *RouterSuite) TestCalculate_UnmarshalError() {
	resp := s.PostJSON("calculate", `{"flights":[}`)
	s.Equal("can't unmarshal request: models.FlightPathRequest.Flights: [][]string: []string: decode "+
		"slice: expect [ or n, but found }, error found in #10 byte of ...|lights\":[}|..., bigger context "+
		"...|{\"flights\":[}|...", resp)
}
