package main

import (
	"fmt"
	"net/http"

	"github.com/modcloth/docker-builder/builder"
	"github.com/modcloth/docker-builder/webhook"

	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/modcloth/auth"
)

//ServerDescription is the help text for the `serer` command
const ServerDescription = `Start a small HTTP web server for receiving build requests.

Configure through the environment:

DOCKER_BUILDER_LOGLEVEL     =>     --log-level (global)
DOCKER_BUILDER_LOGFORMAT    =>     --log-format (global)
DOCKER_BUILDER_PORT         =>     --port
DOCKER_BUILDER_APITOKEN     =>     --api-token
DOCKER_BUILDER_SKIPPUSH     =>     --skip-push
DOCKER_BUILDER_USERNAME     =>     --username
DOCKER_BUILDER_PASSWORD     =>     --password

NOTE: If username and password are both empty (i.e. not provided), basic auth will not be used.
`

var apiToken string

// define the server
func serve(c *cli.Context) {
	// get vars from env and cli

	// set username and password (in helpers.go)
	setUnAndPwd(c)
	// get port
	port := c.Int("port")
	portString := fmt.Sprintf(":%d", port)
	// get api token
	apiToken = c.String("api-token")
	if apiToken == "" {
		apiToken = config.APIToken
	}
	// get "skip-push" option
	builder.SkipPush = c.Bool("skip-push") || config.SkipPush

	// set up server
	server := martini.Classic()

	// check for basic auth
	basicAuthFunc := func(http.ResponseWriter, *http.Request) {}
	if un != "" && pwd != "" {
		basicAuthFunc = auth.Basic(un, pwd)
	}

	// check for travis and github auth
	travisAuthFunc := func(http.ResponseWriter, *http.Request) {}
	if c.String("travis-token") != "" {
		travisAuthFunc = auth.TravisCI(c.String("travis-token"))
	}

	githubAuthFunc := func(http.ResponseWriter, *http.Request) {}
	if c.String("github-secret") != "" {
		githubAuthFunc = auth.GitHub(c.String("github-secret"))
	}

	// configure webhook globals
	webhook.Logger(Logger)
	webhook.APIToken(apiToken)

	// establish routes
	server.Get("/health", func() (int, string) { return 200, "200 OK" })
	server.Post("/docker-build", basicAuthFunc, webhook.DockerBuild)
	server.Post("/docker-build/travis", basicAuthFunc, travisAuthFunc, webhook.Travis)
	server.Post("/docker-build/github", basicAuthFunc, githubAuthFunc, webhook.Github)

	// start server
	http.ListenAndServe(portString, server)
}
