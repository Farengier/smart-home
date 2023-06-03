package signal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

var onShutdown []func() error
var isInitialized = false
var exitCh chan os.Signal
var mtx = &sync.Mutex{}
var wg = &sync.WaitGroup{}
var one = &sync.Once{}

func Init() {
	one.Do(func() {
		onShutdown = make([]func() error, 0, 16)
		isInitialized = true

		exitCh = make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
		go callOnShutdown()
	})
}

func Run(fn func()) {
	if !isInitialized {
		return
	}
	wg.Add(1)
	go (func() {
		defer wg.Done()
		fn()
	})()
}

func OnShutdown(fn func() error) {
	if !isInitialized {
		return
	}

	mtx.Lock()
	defer mtx.Unlock()

	onShutdown = append(onShutdown, fn)
}

func Shutdown() {
	exitCh <- os.Interrupt
}

func Wait() {
	wg.Wait()
}

func callOnShutdown() {
	logrus.Info("[Signal] Waiting stop signal")

	<-exitCh
	mtx.Lock()
	defer mtx.Unlock()

	logrus.Info("[Signal] Stop signal received, shutdown initiated")
	for _, f := range onShutdown {
		f := f
		go (func() {
			_ = f()
		})()
	}
}
