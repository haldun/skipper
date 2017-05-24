package flowid

import (
	"fmt"
	"github.com/zalando/skipper/filters"
	"log"
	"strings"
)

const (
	Name                = "flowID"
	ReuseParameterValue = "reuse"
	HeaderName          = "X-Flow-Id"
)

// Generator interface should be implemented by types that can generate request tracing Flow IDs.
type Generator interface {
	// Generate returns a new Flow ID using the implementation specific format or an error in case of failure.
	Generate() (string, error)
	// MustGenerate behaves like Generate but panics on failure instead of returning an error.
	MustGenerate() string
	// IsValid asserts if a given flow ID follows an expected format
	IsValid(string) bool
}

type flowIDSpec struct {
	generator Generator
}

type flowID struct {
	reuseExisting bool
	generator     Generator
}

// NewFlowID creates a new standard generator with the defined length and returns a Flow ID.
//
// Deprecated: For backward compatibility this exported function is still available but will removed in upcoming
// releases. Use the new Generator interface and respective implementations
func NewFlowID(l int) (string, error) {
	g, err := NewStandardGenerator(l)
	if err != nil {
		return "", fmt.Errorf("deprecated new flowid: %v", err)
	}
	return g.Generate()
}

// New creates a new instance of the flowID filter spec which uses the StandardGenerator.
// To use another type of Generator use NewWithGenerator()
func New() filters.Spec {
	g, err := NewStandardGenerator(defaultLen)
	if err != nil {
		panic(err)
	}
	return NewWithGenerator(g)
}

// New behaves like New but allows you to specify any other Generator.
func NewWithGenerator(g Generator) filters.Spec {
	return &flowIDSpec{generator: g}
}

// Request will inspect the current Request for the presence of an X-Flow-Id header which will be kept in case the
// "reuse" flag has been set. In any other case it will set the same header with the value returned from the
// defined Flow ID Generator
func (f *flowID) Request(fc filters.FilterContext) {
	r := fc.Request()
	var flowID string

	if f.reuseExisting {
		flowID = r.Header.Get(HeaderName)
		if f.generator.IsValid(flowID) {
			return
		}
	}

	flowID, err := f.generator.Generate()
	if err == nil {
		r.Header.Set(HeaderName, flowID)
	} else {
		log.Println(err)
	}
}

// Response is No-Op in this filter
func (*flowID) Response(filters.FilterContext) {}

// CreateFilter will return a new flowID filter from the spec
// If at least 1 argument is present and it contains the value "reuse", the filter instance is configured to accept
// keep the value of the X-Flow-Id header, if it's already set
func (spec *flowIDSpec) CreateFilter(fc []interface{}) (filters.Filter, error) {
	var reuseExisting bool
	if len(fc) > 0 {
		if r, ok := fc[0].(string); ok {
			reuseExisting = strings.ToLower(r) == ReuseParameterValue
		} else {
			return nil, filters.ErrInvalidFilterParameters
		}
		if len(fc) > 1 {
			log.Println("flow id filter warning: this syntaxt is deprecated and will be removed soon. " +
				"please check updated docs")
		}
	}
	return &flowID{reuseExisting: reuseExisting, generator: spec.generator}, nil
}

// Name returns the canonical filter name
func (*flowIDSpec) Name() string { return Name }
