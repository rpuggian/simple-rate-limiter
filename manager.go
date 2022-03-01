package ratelimiter

import (
	"errors"
	"log"
	"sync/atomic"
)

var (
	ErrInvalidLimit           = errors.New("limit must be greater than zero")
	ErrTokenFactoryNotDefined = errors.New("token factory must be defined")
)

// Manager implements a rate limiter interface. Internally it contains
// the in / out / release channels and the current rate limiter state.
type Manager struct {
	errorChan chan error
	//for synchronizing the release of tokens.
	releaseChan chan *Token
	outChan     chan *Token
	inChan      chan struct{}
	//a counter representing the number of pending or waiting token requests.
	needToken    int64
	activeTokens map[string]*Token
	//represents the maximum number of tokens that can be active and is used to determine whether a token can be generated or not.
	limit     int
	makeToken tokenFactory
}

// Acquire is called to acquire a new token
func (m *Manager) Acquire() (*Token, error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case t := <-m.outChan:
		return t, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

// Release is called to release an active token
func (m *Manager) Release(t *Token) {
	go func() {
		m.releaseChan <- t
	}()

}

// NewManager creates a manager type
func NewManager(conf *Config) *Manager {
	m := &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		releaseChan:  make(chan *Token),
		needToken:    0,
		limit:        conf.Limit,
		makeToken:    NewToken,
	}

	return m
}

func (m *Manager) incNeedToken() {
	atomic.AddInt64(&m.needToken, 1)
}

func (m *Manager) decNeedToken() {
	atomic.AddInt64(&m.needToken, -1)
}

func (m *Manager) awaitingToken() bool {
	return atomic.LoadInt64(&m.needToken) > 0
}

func (m *Manager) tryGenerateToken() {
	// panic if token factory is not defined
	if m.makeToken == nil {
		panic(ErrTokenFactoryNotDefined)
	}

	// cannot continue if limit has been reached
	if m.isLimitExceeded() {
		m.incNeedToken()
		return
	}

	token := m.makeToken()

	// Add token to active map
	m.activeTokens[token.ID] = token

	// send token to outChan
	go func() {
		m.outChan <- token
	}()
}

func (m *Manager) isLimitExceeded() bool {
	if len(m.activeTokens) >= m.limit {
		return true
	}
	return false
}

func (m *Manager) releaseToken(token *Token) {
	if token == nil {
		log.Print("unable to relase nil token")
		return
	}

	if _, ok := m.activeTokens[token.ID]; !ok {
		log.Printf("unable to relase token %s - not in use", token)
		return
	}

	// Delete from map
	delete(m.activeTokens, token.ID)

	// process anything waiting for a rate limit
	if m.awaitingToken() {
		m.decNeedToken()
		go m.tryGenerateToken()
	}
}
