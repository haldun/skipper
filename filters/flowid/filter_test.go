package flowid

import (
	"bytes"
	"fmt"
	"github.com/zalando/skipper/filters"
	"github.com/zalando/skipper/filters/filtertest"
	"log"
	"net/http"
	"strings"
	"testing"
)

const (
	testFlowID    = "FLOW-ID-FOR-TESTING"
	invalidFlowID = "[<>] (o) [<>]"
)

var (
	testFlowIDSpec           = New()
	filterConfigWithReuse    = []interface{}{ReuseParameterValue}
	filterConfigWithoutReuse = []interface{}{"dummy"}
)

func TestNewFlowIdGeneration(t *testing.T) {
	f, _ := testFlowIDSpec.CreateFilter(filterConfigWithReuse)
	fc := buildfilterContext()
	f.Request(fc)

	flowID := fc.Request().Header.Get(HeaderName)
	if flowID == "" {
		t.Error("flowID not generated")
	}
}

func TestFlowIdReuseExisting(t *testing.T) {
	f, _ := testFlowIDSpec.CreateFilter(filterConfigWithReuse)
	fc := buildfilterContext(HeaderName, testFlowID)
	f.Request(fc)

	flowID := fc.Request().Header.Get(HeaderName)
	if flowID != testFlowID {
		t.Errorf("Got wrong flow id. Expected '%s' got '%s'", testFlowID, flowID)
	}
}

func TestFlowIdIgnoreReuseExisting(t *testing.T) {
	f, _ := testFlowIDSpec.CreateFilter(filterConfigWithoutReuse)
	fc := buildfilterContext(HeaderName, testFlowID)
	f.Request(fc)

	flowID := fc.Request().Header.Get(HeaderName)
	if flowID == testFlowID {
		t.Errorf("Got wrong flow id. Expected a newly generated flowid but got the test flow id '%s'", flowID)
	}
}

func TestFlowIdRejectInvalidReusedFlowId(t *testing.T) {
	f, _ := testFlowIDSpec.CreateFilter(filterConfigWithReuse)
	fc := buildfilterContext(HeaderName, invalidFlowID)
	f.Request(fc)

	flowID := fc.Request().Header.Get(HeaderName)
	if flowID == invalidFlowID {
		t.Errorf("Got wrong flow id. Expected a newly generated flowid but got the test flow id '%s'", flowID)
	}
}

func TestFlowIdWithInvalidParameters(t *testing.T) {
	fc := []interface{}{true}
	_, err := testFlowIDSpec.CreateFilter(fc)
	if err != filters.ErrInvalidFilterParameters {
		t.Errorf("expected an invalid parameters error, got %v", err)
	}

	fc = []interface{}{1}
	_, err = testFlowIDSpec.CreateFilter(fc)
	if err != filters.ErrInvalidFilterParameters {
		t.Errorf("expected an invalid parameters error, got %v", err)
	}
}

func TestDeprecationNotice(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	fc := []interface{}{"", true}
	_, err := testFlowIDSpec.CreateFilter(fc)
	if err == filters.ErrInvalidFilterParameters {
		t.Error("unexpected error creating a filter")
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("no warning output produced")
	}

	if !(strings.Contains(logOutput, "warning") || strings.Contains(logOutput, "deprecated")) {
		t.Error("missing deprecation keywords from the output produced")
	}
}

func TestFlowIdWithCustomGenerators(t *testing.T) {
	for _, test := range []struct {
		generatorID string
	}{
		{""},
		{"builtin"},
		{"ulid"},
	} {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			fc := []interface{}{ReuseParameterValue, test.generatorID}
			f, _ := testFlowIDSpec.CreateFilter(fc)
			fctx := buildfilterContext()
			f.Request(fctx)

			flowID := fctx.Request().Header.Get(HeaderName)

			l := len(flowID)
			if l == 0 {
				t.Errorf("wrong flowID len. expected > 0 got %d / %q", l, flowID)
			}

		})
	}
}

func buildfilterContext(headers ...string) filters.FilterContext {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	for i := 0; i < len(headers); i += 2 {
		r.Header.Set(headers[i], headers[i+1])
	}
	return &filtertest.Context{FRequest: r}
}
