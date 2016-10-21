package extension_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/extension"
	"github.com/pivotal-cf/go-pivnet/extension/extensionfakes"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
)

var _ = Describe("ExtendedClient", func() {
	var (
		server     *ghttp.Server
		fakeLogger *loggerfakes.FakeLogger
		client     extension.ExtendedClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		token := "my-auth-token"
		userAgent := "go-pivnet/0.1.0"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig := pivnet.ClientConfig{
			Host:      server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		c := pivnet.NewClient(newClientConfig, fakeLogger)
		client = extension.NewExtendedClient(c, fakeLogger)
	})

	Describe("DownloadFile", func() {
		var (
			downloadLink string

			fileContents []byte

			httpStatus int
		)

		BeforeEach(func() {
			downloadLink = "/some/download/link"

			fileContents = []byte("some file contents")

			httpStatus = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s%s",
						apiPrefix,
						downloadLink,
					)),
					ghttp.RespondWith(httpStatus, fileContents),
				),
			)
		})

		It("writes file contents to provided writer", func() {
			writer := bytes.NewBuffer(nil)

			err, retryable := client.DownloadFile(writer, downloadLink)
			Expect(err).NotTo(HaveOccurred())
			Expect(retryable).To(BeTrue())

			Expect(writer.Bytes()).To(Equal(fileContents))
		})

		Context("when creating the request returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some request error")
				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(nil, expectedErr)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err, _ := client.DownloadFile(nil, downloadLink)
				Expect(err).To(Equal(expectedErr))
			})

			It("is not retryable", func() {
				_, retryable := client.DownloadFile(nil, downloadLink)
				Expect(retryable).To(BeFalse())
			})
		})

		Context("when dumping the request returns an error", func() {
			BeforeEach(func() {
				u, err := url.Parse("https://example.com")
				Expect(err).NotTo(HaveOccurred())

				request := &http.Request{
					URL: u,
				}

				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(request, nil)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err, _ := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())
			})

			It("is not retryable", func() {
				_, retryable := client.DownloadFile(nil, downloadLink)
				Expect(retryable).To(BeFalse())
			})
		})

		Context("when making the request returns an error", func() {
			BeforeEach(func() {
				u, err := url.Parse("https://not-a-real-site-5463456.com")
				Expect(err).NotTo(HaveOccurred())

				request := &http.Request{
					Header: http.Header{},
					URL:    u,
				}

				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(request, nil)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err, _ := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())
			})

			It("is not retryable", func() {
				_, retryable := client.DownloadFile(nil, downloadLink)
				Expect(retryable).To(BeFalse())
			})
		})

		Context("when the response status code is 451", func() {
			BeforeEach(func() {
				httpStatus = http.StatusUnavailableForLegalReasons
			})

			It("returns an error", func() {
				err, _ := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("EULA"))
			})

			It("is not retryable", func() {
				_, retryable := client.DownloadFile(nil, downloadLink)
				Expect(retryable).To(BeFalse())
			})
		})

		Context("when the response status code is not 200", func() {
			BeforeEach(func() {
				httpStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				err, _ := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("418"))
			})

			It("is not retryable", func() {
				_, retryable := client.DownloadFile(nil, downloadLink)
				Expect(retryable).To(BeFalse())
			})
		})

		Context("when there is an error copying the contents", func() {
			var (
				writer errWriter
			)

			BeforeEach(func() {
				writer = errWriter{}
			})

			It("returns an error", func() {
				err, _ := client.DownloadFile(writer, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("error writing"))
			})

			It("is retryable", func() {
				_, retryable := client.DownloadFile(writer, downloadLink)
				Expect(retryable).To(BeTrue())
			})
		})
	})
})

type errWriter struct {
}

func (e errWriter) Write([]byte) (int, error) {
	return 0, errors.New("error writing")
}
