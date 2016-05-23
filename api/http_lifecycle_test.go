package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/suite"
)

var (
	root, _ = url.Parse("https://example.com")
)

type HttpLifecycleTestSuite struct {
	suite.Suite
}

func (suite *HttpLifecycleTestSuite) SetupTest() {
	SetupTestCredentialsFunc()
}

func (suite *HttpLifecycleTestSuite) TearDownTest() {
	RestoreCredentialsFunc()
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleMakesRequestsAgainstAbsolutePath() {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
	})

	suite.Assert().Nil(err)
	suite.Assert().Equal("https://example.com/foo", req.URL.String())
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleAttachesQueryParameters() {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
		Query: map[string]string{
			"a": "b",
		},
	})

	suite.Assert().Nil(err)
	suite.Assert().Equal("https://example.com/foo?a=b", req.URL.String())
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleAttachesBodyWhenPresent() {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{
		Body: struct {
			Foo string `json:"foo"`
		}{"bar"},
	})

	suite.Assert().Nil(err)

	body, err := ioutil.ReadAll(req.Body)
	suite.Assert().Nil(err)
	suite.Assert().Equal("{\"foo\":\"bar\"}", string(body))
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleDoesNotAttachBodyWhenEmpty() {
	l := api.NewHttpLifecycle(root)
	req, err := l.Build(&api.RequestSchema{})

	suite.Assert().Nil(err)
	suite.Assert().Nil(req.Body)
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleExecutesRequestWithoutBody() {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		suite.Assert().Equal("/path", r.URL.RequestURI())
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/path", nil)

	l := api.NewHttpLifecycle(root)
	_, err := l.Execute(req, nil)

	suite.Assert().True(called)
	suite.Assert().Nil(err)
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleExecutesRequestWithBody() {
	type Response struct {
		Foo string `json:"foo"`
	}

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		w.Write([]byte("{\"foo\":\"bar\"}"))
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/path", nil)

	l := api.NewHttpLifecycle(root)
	resp := new(Response)
	_, err := l.Execute(req, resp)

	suite.Assert().True(called)
	suite.Assert().Nil(err)
	suite.Assert().Equal("bar", resp.Foo)
}
