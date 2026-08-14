// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package wsconn

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thediveo/safe" // ahh, now I feel so safe, finally...

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gcustom"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success" // fängschui for devs
	. "github.com/thediveo/testily/concur"
	. "github.com/thediveo/testily/tuples"
)

// BeACloseError is a custom matcher that succeeds if actual is a
// *websocket.CloseError with the specified code and text; otherwise, it fails.
func BeACloseError(code int, text string) types.GomegaMatcher {
	return MakeMatcher(func(err *websocket.CloseError) (bool, error) {
		return err.Code == code && err.Text == text, nil
	}).WithTemplate("Expected:\n{{.FormattedActual}}\n{{.To}} be a CloseError with code {{.Data.Code}} and Text {{.Data.Text}}",
		struct {
			Code int
			Text string
		}{Code: code, Text: strconv.Quote(text)})
}

func wsurl(url string) string { return "ws://" + strings.TrimPrefix(url, "http://") }

var _ = Describe("web socket connections", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	var (
		log *safe.Buffer
		url string
		ch  chan Pair[*WSConn, error]
	)

	BeforeEach(func() {
		log = &safe.Buffer{}
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(log, GinkgoWriter),
			&slog.HandlerOptions{Level: slog.LevelDebug})))

		By("creating a test server with an upgrading websocket handler")
		ch = make(chan Pair[*WSConn, error], 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ch <- PackPair(NewWSConn(w, req))
			// one way or the other we shouldn't touch the response writer
			// directly anymore; and in particular not in case of error, as the
			// upgrade will already have done its dirty response work.
		}))
		url = server.URL
		DeferCleanup(func() {
			server.CloseClientConnections()
			server.Close()
		})
	})

	It("rejects and logs an invalid connection attempt", func() {
		By("connecting")
		resp := Successful(http.Get(url))
		DeferCleanup(resp.Body.Close)
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		Eventually(log.String).To(MatchRegexp(
			`level=ERROR msg="websocket upgrade process failed" connection=[a-z]+-[a-z]+`))
	})

	It("successfully creates a new socket connection wrapper", func(ctx context.Context) {
		By("connecting")
		conn, resp := Successful2R(
			websocket.DefaultDialer.DialContext(ctx, wsurl(url), nil))
		DeferCleanup(resp.Body.Close)
		DeferCleanup(conn.Close)

		By("checking the WSConn")
		var res Pair[*WSConn, error]
		Eventually(ch).Should(Receive(&res))
		wsconn := Successful(res.Unpack())
		Expect(wsconn).NotTo(BeNil())
		Expect(wsconn.Close()).To(Succeed())

		Expect(log.String()).To(BeEmpty())
	})

	It("watches the connection and gracefully closes when the client requests it", func(ctx context.Context) {
		By("connecting")
		clntconn, resp := Successful2R(
			websocket.DefaultDialer.DialContext(ctx, wsurl(url), nil))
		DeferCleanup(resp.Body.Close)
		DeferCleanup(clntconn.Close)

		By("checking the WSConn")
		var res Pair[*WSConn, error]
		Eventually(ch).Should(Receive(&res))
		wsconn := Successful(res.Unpack())

		By("watching the connection on the server side")
		done := CloseWhenGone(wsconn.Watch)
		DeferCleanup(wsconn.Close)
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg="monitoring.*started"`))

		By("sending a close message to the server")
		Expect(clntconn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "it's fine"))).To(Succeed())
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg=".*started close sequence"`))

		By("receiving the correct close response")
		Expect(clntconn.SetReadDeadline(time.Now().Add(5 * time.Second))).To(Succeed())
		_, _, err := clntconn.ReadMessage()
		Expect(err).To(BeACloseError(websocket.CloseNormalClosure, "ciao"))
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg="monitoring .* ended"`))

		Eventually(done).Should(BeClosed())

		By("trying to watch the closed connection")
		done = CloseWhenGone(wsconn.Watch)
		Eventually(done).Should(BeClosed())
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg="won't monitor .* closed websocket connection"`))

		By("trying to initiate a close on the closed connection")
		wsconn.InitiateGracefulClose(websocket.CloseGoingAway, "bye-bye")
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg=".* already .* closed"`))
	})

	It("gracefully closes its connection to the client", func(ctx context.Context) {
		By("connecting")
		clntconn, resp := Successful2R(
			websocket.DefaultDialer.DialContext(ctx, wsurl(url), nil))
		DeferCleanup(resp.Body.Close)
		DeferCleanup(clntconn.Close)

		By("checking the WSConn")
		var res Pair[*WSConn, error]
		Eventually(ch).Should(Receive(&res))
		wsconn := Successful(res.Unpack())

		// Note: don't watch the server side, as we want to exercise GracefullyClose

		By("watching the connection on the client side")
		go func() {
			defer GinkgoRecover()
			for {
				_, _, err := clntconn.ReadMessage()
				Expect(err).To(BeACloseError(websocket.CloseGoingAway, "bye-bye"))
				return
			}
		}()

		By("issuing a close message to the client")
		closeDone := CloseWhenGone(func() {
			wsconn.GracefullyClose(websocket.CloseGoingAway, "bye-bye")
		})
		Eventually(closeDone).Should(BeClosed())

		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`level=DEBUG msg="monitoring .* ended"`))
	})

	It("gracefully handles a closed close", func(ctx context.Context) {
		By("connecting")
		clntconn, resp := Successful2R(
			websocket.DefaultDialer.DialContext(ctx, wsurl(url), nil))
		DeferCleanup(resp.Body.Close)
		DeferCleanup(clntconn.Close)

		By("checking the WSConn")
		var res Pair[*WSConn, error]
		Eventually(ch).Should(Receive(&res))
		wsconn := Successful(res.Unpack())

		By("dropping the connection on the client side")
		Expect(clntconn.Close()).To(Succeed())

		By("attempting to close the closed connection")
		closeDone := CloseWhenGone(func() {
			wsconn.GracefullyClose(websocket.CloseGoingAway, "")
		})
		Eventually(closeDone).Should(BeClosed())
		Eventually(log.String).Within(5 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(MatchRegexp(`(?m)level=DEBUG .* code=1006 .*\n([^\n]+\n)+.* level=DEBUG msg="monitoring .* ended"`))
	})

})
