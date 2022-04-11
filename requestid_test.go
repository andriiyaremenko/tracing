package tracing_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/andriiyaremenko/tracing"
)

var _ = Describe("RequestID", func() {
	getIDConstructor := func() func() string {
		i := 0
		return func() string {
			defer func() { i++ }()

			return strconv.Itoa(i)
		}
	}
	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
	})
	requestHandler := func(callback func(r *http.Request)) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callback(r)
			_, _ = io.Copy(w, r.Body)
		})
	}
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	getID := getIDConstructor()
	middleware := tracing.Middleware(tracing.DefaultRequestIDOptions, getID)

	Context("should return new ID if one was not found in headers", func() {
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/echo", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"0"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/error", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						requestID, ok := tracing.GetTracing[tracing.RequestID](r.Context())

						Expect(ok).To(BeTrue())
						Expect(requestID).To(Equal(tracing.RequestID("2")))

					}),
				),
			)

			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/request", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()
		})
	})

	Context("should return same ID if one was found in headers", func() {
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		getRequest := func(addr string) *http.Request {
			r, _ := http.NewRequest(http.MethodGet, addr, nil)

			tracing.DefaultRequestIDWriteHeader(r.Header, "1")

			return r
		}

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/echo", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"1"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/error", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						requestID, ok := tracing.GetTracing[tracing.RequestID](r.Context())

						Expect(ok).To(BeTrue())
						Expect(requestID).To(Equal(tracing.RequestID("1")))
					}),
				),
			)

			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/request", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()
		})
	})

	Context("should use RequestID header name from options if provided", func() {
		fromHeader := tracing.RequestIDReadHeader("X-My-Request-Id")
		setHeader := tracing.RequestIDWriteHeader("X-My-Request-Id")
		middleware := tracing.Middleware(tracing.RequestIDOptions(fromHeader, setHeader), getID)
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		getRequest := func(addr string) *http.Request {
			r, _ := http.NewRequest(http.MethodGet, addr, nil)

			setHeader(r.Header, "1")

			return r
		}

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/echo", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue("X-My-Request-Id", []string{"1"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/error", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue("X-My-Request-Id", []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						requestID, ok := tracing.GetTracing[tracing.RequestID](r.Context())

						Expect(ok).To(BeTrue())
						Expect(requestID).To(Equal(tracing.RequestID("1")))
					}),
				),
			)

			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/request", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()
		})
	})
})
