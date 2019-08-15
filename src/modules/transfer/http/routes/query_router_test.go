package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/backend"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/config"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/http/middleware"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/http/render"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

type TestCase struct {
	method     string
	url        string
	path       string
	body       interface{}
	expectBody string
	handler    http.HandlerFunc
}

func TestTsdbDrawDataQuery(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	err := config.ParseConfig("../../../../../etc/transfer.yml")
	if err != nil {
		log.Fatalln(err)
	}

	backend.Init()

	render.Init()

	tc := TestCase{
		method:     "POST",
		url:        "/api/v1/data",
		path:       "/api/v1/data",
		body:       dataobj.NewQueryData(),
		expectBody: `{"dat":[],"err":""}`,
		handler:    TsdbDrawDataQuery,
	}
	CheckResp(tc, t)
}

func CheckResp(test TestCase, t *testing.T) {
	b, _ := json.Marshal(test.body)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(test.method, test.url, strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	r := mux.NewRouter()
	n := negroni.New()

	n.Use(middleware.NewRecovery())
	r.HandleFunc(test.path, test.handler).Methods(test.method)

	n.UseHandler(r)
	n.ServeHTTP(rr, req)

	if rr.Body.String() != test.expectBody {
		t.Errorf("%v handler returned unexpected body: got %v want %v", test.body,
			rr.Body.String(), test.expectBody)
	}
	//log.Println(test, rr.Body.String())
}
