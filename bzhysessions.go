// Copyright 2020 Wayne wang<net_use@bzhy.com>.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bzhysessions

import (
	//	"github.com/wangyysde/bzhylog"
	"net/http"

	"golang/pipe/golang/pkg/mod/github.com/gorilla/sessions@v1.2.0"
	"golang/pipe/golang/pkg/mod/github.com/wangyysde/bzhysessions@v0.0.0-20200903142906-3777de038f25"

	"github.com/wangyysde/bzhycontext"
	"github.com/wangyysde/bzhyserver"
)

const (
	DefaultKey  = "bzhysession"
	errorFormat = "[sessions] ERROR! %s\n"
)

type Store interface {
	bzhysessions.Store
	Options(Options)
}

// Wraps thinly gorilla-session methods.
// Session stores the values and optional configuration for a session.
type Session interface {
	// Get returns the session value associated to the given key.
	Get(key interface{}) interface{}
	// Set sets the session value associated to the given key.
	Set(key interface{}, val interface{})
	// Delete removes the session value associated to the given key.
	Delete(key interface{})
	// Clear deletes all values in the session.
	Clear()
	// AddFlash adds a flash message to the session.
	// A single variadic argument is accepted, and it is optional: it defines the flash key.
	// If not defined "_flash" is used by default.
	AddFlash(value interface{}, vars ...string)
	// Flashes returns a slice of flash messages from the session.
	// A single variadic argument is accepted, and it is optional: it defines the flash key.
	// If not defined "_flash" is used by default.
	Flashes(vars ...string) []interface{}
	// Options sets configuration for a session.
	Options(Options)
	// Save saves all sessions used during the current request.
	Save() error
}

func Sessions(name string, store Store) bzhyserver.HandlerFunc {
	return func(c *bzhyserver.Context) {
		s := &session{name, c.Request, store, nil, false, c.Writer}
		c.Set(DefaultKey, s)
		defer bzhycontext.Clear(c.Request)
		c.Next()
	}
}

func SessionsMany(names []string, store Store) bzhyserver.HandlerFunc {
	return func(c *bzhyserver.Context) {
		sessions := make(map[string]Session, len(names))
		for _, name := range names {
			sessions[name] = &session{name, c.Request, store, nil, false, c.Writer}
		}
		c.Set(DefaultKey, sessions)
		defer bzhycontext.Clear(c.Request)
		c.Next()
	}
}

type session struct {
	name    string
	request *http.Request
	store   Store
	session *bzhysessions.Session
	written bool
	writer  http.ResponseWriter
}

func (s *session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

func (s *session) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

func (s *session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

func (s *session) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

func (s *session) Flashes(vars ...string) []interface{} {
	s.written = true
	return s.Session().Flashes(vars...)
}

func (s *session) Options(options Options) {
	s.Session().Options = options.ToGorillaOptions()
}

func (s *session) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.writer)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

func (s *session) Session() *sessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.name)
		//		if err != nil {
		//			log.Printf(errorFormat, err)
		//		}
	}
	return s.session
}

func (s *session) Written() bool {
	return s.written
}

// shortcut to get session
func Default(c *bzhyserver.Context) Session {
	return c.MustGet(DefaultKey).(Session)
}

// shortcut to get session with given name
func DefaultMany(c *bzhyserver.Context, name string) Session {
	return c.MustGet(DefaultKey).(map[string]Session)[name]
}
