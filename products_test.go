package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("PivnetClient - product", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      lager.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = lager.NewLogger("products test")
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Get", func() {
		var (
			slug = "my-product"
		)
		Context("when the product can be found", func() {
			It("returns the located product", func() {
				response := fmt.Sprintf(`{"id": 3, "slug": "%s"}`, slug)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				product, err := client.Products.Get(slug)
				Expect(err).NotTo(HaveOccurred())
				Expect(product.Slug).To(Equal(slug))
			})
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.Products.Get(slug)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("List", func() {
		var (
			slug = "my-product"
		)

		Context("when the products can be found", func() {
			It("returns the products", func() {
				response := fmt.Sprintf(`{"products":[{"id": 3, "slug": "%s"},{"id": 4, "slug": "bar"}]}`, slug)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products",
							apiPrefix)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				products, err := client.Products.List()
				Expect(err).NotTo(HaveOccurred())

				Expect(products).To(HaveLen(2))
				Expect(products[0].Slug).To(Equal(slug))
			})
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products",
							apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.Products.List()
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
