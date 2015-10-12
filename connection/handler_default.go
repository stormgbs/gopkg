package connection

import (
	"log"
)

type DataHandler interface {
	ProcessRequest([]byte) ([]byte, error)
}

type ErrorHandler interface {
	OnError(error)
}

type default_data_handler struct{}

func (ddh *default_data_handler) ProcessRequest(data []byte) ([]byte, error) {
	/* do nothing */
	return nil, nil
}

type default_error_handler struct{}

func (deh *default_error_handler) OnError(err error) {
	log.Printf("OnError(): %s", err)
}
