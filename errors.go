package wera

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	ErrCloseConn = errors.New("close connection")
	ErrNoPubSub  = errors.New("pub-sub implementation is not provided")
)

var expectedCloseErrors = []int{
	websocket.CloseNormalClosure,
	websocket.CloseGoingAway,
	websocket.CloseProtocolError,
	websocket.CloseUnsupportedData,
	websocket.CloseNoStatusReceived,
	websocket.CloseAbnormalClosure,
	websocket.CloseInvalidFramePayloadData,
	websocket.ClosePolicyViolation,
	websocket.CloseMessageTooBig,
	websocket.CloseMandatoryExtension,
	websocket.CloseInternalServerErr,
	websocket.CloseServiceRestart,
	websocket.CloseTryAgainLater,
	websocket.CloseTLSHandshake,
}

func fmtCloseErr(internalErr error) error {
	return fmt.Errorf("%w: %s", ErrCloseConn, internalErr)
}
