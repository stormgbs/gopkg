package connection

import (
	"errors"
	"log"
)

var ErrOrphanRespDiscard = errors.New("discard orphan response")

type DataHandler interface {
	//handle request
	ProcessRequest([]byte) ([]byte, error)

	//only handle response which found no app associated.
	ProcessOrphanResponse([]byte) error
}

type ErrorHandler interface {
	OnError(error)
}

type default_data_handler struct{}

func (ddh *default_data_handler) ProcessRequest(data []byte) ([]byte, error) {
	/* do nothing */
	return nil, nil
}

func (ddh *default_data_handler) ProcessOrphanResponse(data []byte) error {
	/* do nothing */
	return ErrOrphanRespDiscard
}

type default_error_handler struct{}

func (deh *default_error_handler) OnError(err error) {
	log.Printf("OnError(): %v", err)
}
