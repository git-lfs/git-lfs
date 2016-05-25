package api_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/git-lfs/api"
	"github.com/github/git-lfs/config"
	"github.com/stretchr/testify/suite"
)

type NopEndpointSource struct {
	Root string
}

func (e *NopEndpointSource) Endpoint(op string) config.Endpoint {
	return config.Endpoint{Url: e.Root}
}

var (
	source = &NopEndpointSource{"https://example.com"}
)

func TestHttpLifecycleSuite(t *testing.T) {
	suite.Run(t, new(HttpLifecycleTestSuite))
}

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
	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path:      "/foo",
		Operation: api.DownloadOperation,
	})

	suite.Assert().Nil(err)
	suite.Assert().Equal("https://example.com/foo", req.URL.String())
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleAttachesQueryParameters() {
	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path:      "/foo",
		Operation: api.DownloadOperation,
		Query: map[string]string{
			"a": "b",
		},
	})

	suite.Assert().Nil(err)
	suite.Assert().Equal("https://example.com/foo?a=b", req.URL.String())
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleAttachesBodyWhenPresent() {
	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Operation: api.DownloadOperation,
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
	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Operation: api.DownloadOperation,
	})

	suite.Assert().Nil(err)
	suite.Assert().Nil(req.Body)
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleErrsWithoutOperation() {
	l := api.NewHttpLifecycle(source)
	req, err := l.Build(&api.RequestSchema{
		Path: "/foo",
	})

	suite.Assert().Equal(api.ErrNoOperationGiven, err)
	suite.Assert().Nil(req)
}

func (suite *HttpLifecycleTestSuite) TestHttpLifecycleExecutesRequestWithoutBody() {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		suite.Assert().Equal("/path", r.URL.RequestURI())
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/path", nil)

	l := api.NewHttpLifecycle(source)
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

	req, _ := http.NewRequest("GET", server.URL+"/path", nil)

	l := api.NewHttpLifecycle(source)
	resp := new(Response)
	_, err := l.Execute(req, resp)

	suite.Assert().True(called)
	suite.Assert().Nil(err)
	suite.Assert().Equal("bar", resp.Foo)
}
