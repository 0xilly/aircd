// SPDX-License-Identifier: BSD-2-Clause
package transport

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0xilly/aircd/ircd/internal/core"
	"github.com/0xilly/aircd/ircd/internal/irc"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	ws     *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	limit   int // max IRC line bytes
	pending strings.Builder
	mu      sync.Mutex // guards WriteMessage (Send)
}

func (w *wsConn) Context() context.Context { return w.ctx }
func (w *wsConn) Addr() string             { return w.ws.RemoteAddr().String() }
func (w *wsConn) Close() error {
	w.cancel()
	return w.ws.Close()
}

// ReadLine returns the next IRC line (no CRLF). It may assemble it from multiple WS frames.
func (w *wsConn) ReadLine() (string, error) {
	for {
		// Check if we already buffered a full line
		if idx := strings.IndexByte(w.pending.String(), '\n'); idx >= 0 {
			line := w.pending.String()[:idx]
			rest := w.pending.String()[idx+1:]
			w.pending.Reset()
			w.pending.WriteString(rest)

			line = strings.TrimRight(line, "\r")
			if len(line) > w.limit {
				line = line[:w.limit]
			}
			return line, nil
		}

		// Need more data: read next text frame
		msgType, data, err := w.ws.ReadMessage()
		if err != nil {
			return "", err
		}
		if msgType != websocket.TextMessage {
			// Ignore non-text frames (binary, ping/pong handled internally)
			continue
		}
		// Append; may contain multiple lines
		w.pending.Write(data)
		// Normalize CRLF? We just search for '\n' above; CRs are trimmed later.
	}
}

func (w *wsConn) Send(m *irc.Message) error {
	// One IRC message per WS text frame.
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ws.WriteMessage(websocket.TextMessage, []byte(m.String()+"\r\n"))
}

type TLSConfig interface {
	Enabled() bool
	CertFile() string
	KeyFile() string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: tighten this
		return true
	},
}

func ServeWS(srv *core.Server, maxLine int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Optional: tune deadlines & ping/pong
		ws.SetReadLimit(int64(maxLine) * 4)
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())
		conn := &wsConn{
			ws:     ws,
			ctx:    ctx,
			cancel: cancel,
			limit:  maxLine,
		}

		// Writer ping loop (optional)
		go func() {
			t := time.NewTicker(30 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					conn.mu.Lock()
					_ = ws.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
					conn.mu.Unlock()
				}
			}
		}()

		srv.Accept(conn)
	})
}

func ListenWS(addr, path string, srv *core.Server, maxLine int, tlsEnabled bool, cert, key string) error {
	mux := http.NewServeMux()
	mux.Handle(path, ServeWS(srv, maxLine))

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if tlsEnabled {
		return httpSrv.ListenAndServeTLS(cert, key)
	}
	return httpSrv.ListenAndServe()
}
