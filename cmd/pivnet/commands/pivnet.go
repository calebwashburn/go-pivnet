package commands

import (
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
	"github.com/pivotal-golang/lager"
)

const (
	printAsTable = "table"
	printAsJSON  = "json"
	printAsYAML  = "yaml"

	defaultHost = "https://network.pivotal.io"
)

type PivnetCommand struct {
	Version func() `short:"v" long:"version" description:"Print the version of Pivnet and exit"`

	Format string `long:"format" description:"Format to print as" default:"table" choice:"table" choice:"json" choice:"yaml"`

	APIToken string `long:"api-token" description:"Pivnet API token"`
	Host     string `long:"host" description:"Pivnet API Host"`

	ReleaseTypes ReleaseTypesCommand `command:"release-types" description:"List release types"`

	EULAs      EULAsCommand      `command:"eulas" description:"List EULAs"`
	AcceptEULA AcceptEULACommand `command:"accept-eula" description:"Accept EULA"`

	Products ProductsCommand `command:"products" description:"List products"`
	Product  ProductCommand  `command:"product" description:"Show product"`

	Releases      ReleasesCommand      `command:"releases" description:"List releases"`
	Release       ReleaseCommand       `command:"release" description:"Show release"`
	DeleteRelease DeleteReleaseCommand `command:"delete-release" description:"Delete release"`

	UserGroups UserGroupsCommand `command:"user-groups" description:"List user groups"`

	ReleaseDependencies ReleaseDependenciesCommand `command:"release-dependencies" description:"List user groups"`
}

var Pivnet PivnetCommand

func init() {
	Pivnet.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if Pivnet.Host == "" {
		Pivnet.Host = defaultHost
	}
}

func NewClient() pivnet.Client {
	useragent := fmt.Sprintf(
		"go-pivnet/%s",
		version.Version,
	)

	pivnetClient := pivnet.NewClient(
		pivnet.ClientConfig{
			Token:     Pivnet.APIToken,
			Host:      Pivnet.Host,
			UserAgent: useragent,
		},
		lager.NewLogger("pivnet CLI"),
	)

	return pivnetClient
}
