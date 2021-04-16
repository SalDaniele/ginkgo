package internal

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/onsi/ginkgo/formatter"
)

type InterruptStatus struct {
	Interrupted bool
	Channel     chan interface{}
}

type InterruptHandlerInterface interface {
	Status() InterruptStatus
	SetInterruptMessage(string)
	ClearInterruptMessage()
}

type InterruptHandler struct {
	c                chan interface{}
	lock             *sync.Mutex
	interrupted      bool
	interruptMessage string
}

func NewInterruptHandler() *InterruptHandler {
	handler := &InterruptHandler{
		c:           make(chan interface{}),
		lock:        &sync.Mutex{},
		interrupted: false,
	}
	handler.registerForInterrupts()
	return handler
}

func (handler *InterruptHandler) registerForInterrupts() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-c
			handler.lock.Lock()
			if handler.interruptMessage != "" {
				fmt.Println(handler.interruptMessage)
			}
			handler.interrupted = true
			close(handler.c)
			handler.c = make(chan interface{})
			handler.lock.Unlock()
		}
	}()
}

func (handler *InterruptHandler) Status() InterruptStatus {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	return InterruptStatus{
		Interrupted: handler.interrupted,
		Channel:     handler.c,
	}
}

func (handler *InterruptHandler) SetInterruptMessage(message string) {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = message
}

func (handler *InterruptHandler) ClearInterruptMessage() {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	handler.interruptMessage = ""
}

func interruptMessageWithStackTraces() string {
	out := "Interrupted by User\n\n"
	out += "Here's a stack trace of all running goroutines:\n"
	buf := make([]byte, 8192)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	out += formatter.Fi(1, "%s", string(buf))
	return out
}