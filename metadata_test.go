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

var _ = Describe("Metadata", func() {
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
	middleware := tracing.Middleware(tracing.DefaultMetadataOptions, getID)

	Context("should return new if one was not found in headers", func() {
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/echo", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"0"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCausationID, []string{"0"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCorrelationID, []string{"0"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/error", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"1"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCausationID, []string{"1"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCorrelationID, []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						metadata, ok := tracing.GetTracing[tracing.Metadata](r.Context())

						Expect(ok).To(BeTrue())
						Expect(metadata.ID).To(Equal("2"))
						Expect(metadata.CausationID).To(Equal("2"))
						Expect(metadata.CorrelationID).To(Equal("2"))
					}),
				),
			)

			server := httptest.NewServer(mux)
			resp, err := http.Get(fmt.Sprintf("%s/request", server.URL))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()
		})
	})

	Context("should calculate next metadata if one was found in headers", func() {
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		getRequest := func(addr string) *http.Request {
			r, _ := http.NewRequest(http.MethodGet, addr, nil)

			tracing.DefaultMetadataWriteHeader(
				r.Header,
				tracing.Metadata{ID: "2", CausationID: "1", CorrelationID: "1"},
			)

			return r
		}

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/echo", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"3"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCausationID, []string{"2"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCorrelationID, []string{"1"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/error", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderRequestID, []string{"4"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCausationID, []string{"2"}))
			Expect(resp.Header).To(HaveKeyWithValue(tracing.HeaderCorrelationID, []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						metadata, ok := tracing.GetTracing[tracing.Metadata](r.Context())

						Expect(ok).To(BeTrue())
						Expect(metadata.ID).To(Equal("5"))
						Expect(metadata.CausationID).To(Equal("2"))
						Expect(metadata.CorrelationID).To(Equal("1"))
					}),
				),
			)

			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/request", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()
		})
	})

	Context("should use metadata header names from options if provided", func() {
		fromHeader := tracing.MetadataReadHeader("X-My-Request-Id", "X-My-Causation-Id", "X-My-Correlation-Id")
		setHeader := tracing.MetadataWriteHeader("X-My-Request-Id", "X-My-Causation-Id", "X-My-Correlation-Id")
		middleware := tracing.Middleware(tracing.MetadataOptions(fromHeader, setHeader), getID)
		mux := http.NewServeMux()

		mux.Handle("/echo", middleware(echoHandler))
		mux.Handle("/error", middleware(errorHandler))

		getRequest := func(addr string) *http.Request {
			r, _ := http.NewRequest(http.MethodGet, addr, nil)

			setHeader(
				r.Header,
				tracing.Metadata{ID: "2", CausationID: "1", CorrelationID: "1"},
			)

			return r
		}

		It("for successful response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/echo", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue("X-My-Request-Id", []string{"6"}))
			Expect(resp.Header).To(HaveKeyWithValue("X-My-Causation-Id", []string{"2"}))
			Expect(resp.Header).To(HaveKeyWithValue("X-My-Correlation-Id", []string{"1"}))
		})

		It("for error response", func() {
			server := httptest.NewServer(mux)
			resp, err := http.DefaultClient.Do(getRequest(fmt.Sprintf("%s/error", server.URL)))

			Expect(err).ShouldNot(HaveOccurred())

			resp.Body.Close()

			Expect(resp.Header).To(HaveKeyWithValue("X-My-Request-Id", []string{"7"}))
			Expect(resp.Header).To(HaveKeyWithValue("X-My-Causation-Id", []string{"2"}))
			Expect(resp.Header).To(HaveKeyWithValue("X-My-Correlation-Id", []string{"1"}))
		})

		It("in request context", func() {
			mux.Handle(
				"/request",
				middleware(
					requestHandler(func(r *http.Request) {
						defer GinkgoRecover()

						metadata, ok := tracing.GetTracing[tracing.Metadata](r.Context())

						Expect(ok).To(BeTrue())
						Expect(metadata.ID).To(Equal("8"))
						Expect(metadata.CausationID).To(Equal("2"))
						Expect(metadata.CorrelationID).To(Equal("1"))
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
