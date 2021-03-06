package main

import (
	"errors"
	"testing"
)

var (
	pkg  *Package
	typs map[string]*Type
)

func init() {
	packages := getPackages()

	if len(packages) != 1 {
		err := errors.New("should have only found 1 package")
		panic(err)
	}

	pkg = packages[0]

	typs = make(map[string]*Type)
	for _, typ := range packages[0].Types {
		typs[typ.Name] = typ
	}
}

func TestGenSpecParsing(t *testing.T) {
	dummy := "dummy"

	s1 := `// Here is a description of some type
// gen that may span lines`
	_, found1 := getGenSpec(s1, dummy)

	if found1 {
		t.Errorf("no gen spec should have been found")
	}

	s2 := `// Here is a description of some type
// +gen`
	spec2, found2 := getGenSpec(s2, dummy)

	if !found2 {
		t.Errorf("gen spec should have been found")
	}

	if spec2 == nil {
		t.Errorf("gen spec should not be nil")
	}

	if len(spec2.Pointer) > 0 {
		t.Errorf("gen spec should not be pointer by default")
	}

	if spec2.Methods != nil {
		t.Errorf("gen spec methods should be nil if unspecified")
	}

	if spec2.Projections != nil {
		t.Errorf("gen spec methods should be nil if unspecified")
	}

	s3 := `// Here is a description of some type
// +gen *`
	spec3, found3 := getGenSpec(s3, dummy)

	if !found3 {
		t.Errorf("gen spec should have been found")
	}

	if spec3 == nil {
		t.Errorf("gen spec should not be nil")
	}

	if spec3.Pointer != "*" {
		t.Errorf("gen spec should be pointer")
	}

	s4 := `// Here is a description of some type
// +gen * methods:"Any,All"`
	spec4, found4 := getGenSpec(s4, dummy)

	if !found4 {
		t.Errorf("gen spec should have been found")
	}

	if spec4 == nil {
		t.Errorf("gen spec should not be nil")
	}

	if spec4.Pointer != "*" {
		t.Errorf("gen spec should be pointer")
	}

	if len(spec4.Methods.Items) != 2 {
		t.Errorf("gen spec should have 2 methods")
	}

	if spec4.Projections != nil {
		t.Errorf("gen spec projections should be nil if unspecified")
	}

	s5 := `// Here is a description of some type
// +gen methods:"Any,All" projections:"GroupBy"`
	spec5, found5 := getGenSpec(s5, dummy)

	if !found5 {
		t.Errorf("gen spec should have been found")
	}

	if spec5 == nil {
		t.Errorf("gen spec should not be nil")
	}

	if len(spec5.Pointer) > 0 {
		t.Errorf("gen spec should not be pointer")
	}

	if len(spec5.Methods.Items) != 2 {
		t.Errorf("gen spec should have 2 subsetted methods")
	}

	if len(spec5.Projections.Items) != 1 {
		t.Errorf("gen spec should have 1 projected type")
	}

	s6 := `// Here is a description of some type
// +gen methods:"" projections:""`
	spec6, found6 := getGenSpec(s6, dummy)

	if !found6 {
		t.Errorf("gen spec should have been found")
	}

	if spec6 == nil {
		t.Errorf("gen spec should not be nil")
	}

	if len(spec6.Pointer) > 0 {
		t.Errorf("gen spec should not be pointer")
	}

	if spec6.Methods == nil {
		t.Errorf("gen spec methods should exist even if empty")
	}

	if len(spec6.Methods.Items) > 0 {
		t.Errorf("gen spec methods should be empty, instead got %v", len(spec6.Methods.Items))
	}

	if spec6.Projections == nil {
		t.Errorf("gen spec projections should exist even if empty")
	}

	if len(spec6.Projections.Items) > 0 {
		t.Errorf("gen spec projections should be empty, instead got %v", spec6.Projections.Items)
	}
}

func TestMethodDetermination(t *testing.T) {
	dummy := "dummy"

	spec1 := &GenSpec{"", dummy, nil, nil}

	standardMethods1, projectionMethods1, err1 := determineMethods(spec1)

	if err1 != nil {
		t.Errorf("empty methods should be ok, instead got '%v'", err1)
	}

	if len(standardMethods1) != len(getStandardMethodKeys()) {
		t.Errorf("standard methods should default to all")
	}

	if len(projectionMethods1) != 0 {
		t.Errorf("projection methods without projected type should be none, instead got %v", projectionMethods1)
	}

	spec2 := &GenSpec{"", dummy, &GenTag{[]string{"Count", "Where"}}, nil}

	standardMethods2, projectionMethods2, err2 := determineMethods(spec2)

	if err2 != nil {
		t.Errorf("empty methods should be ok, instead got %v", err2)
	}

	if len(standardMethods2) != 2 {
		t.Errorf("standard methods should be parsed")
	}

	if len(projectionMethods2) != 0 {
		t.Errorf("projection methods without projected typs should be none")
	}

	spec3 := &GenSpec{"", dummy, &GenTag{[]string{"Count", "Unknown"}}, &GenTag{[]string{}}}

	standardMethods3, projectionMethods3, err3 := determineMethods(spec3)

	if err3 == nil {
		t.Errorf("unknown type should be error")
	}

	if len(standardMethods3) != 1 {
		t.Errorf("standard methods should be parsed, minus unknown")
	}

	if len(projectionMethods3) != 0 {
		t.Errorf("projection methods without projected typs should be none")
	}

	spec4 := &GenSpec{"", dummy, nil, &GenTag{[]string{"SomeType"}}}

	standardMethods4, projectionMethods4, err4 := determineMethods(spec4)

	if err4 != nil {
		t.Errorf("projected typs without subsetted methods should be ok, instead got: '%v'", err4)
	}

	if len(standardMethods4) != len(getStandardMethodKeys()) {
		t.Errorf("standard methods should default to all")
	}

	if len(projectionMethods4) != len(getProjectionMethodKeys()) {
		t.Errorf("projection methods should default to all in presence of projected typs")
	}

	spec5 := &GenSpec{"", dummy, &GenTag{[]string{"GroupBy"}}, &GenTag{[]string{"SomeType"}}}

	standardMethods5, projectionMethods5, err5 := determineMethods(spec5)

	if err5 != nil {
		t.Errorf("projected typs with subsetted methods should be ok, instead got: '%v'", err5)
	}

	if len(standardMethods5) != 0 {
		t.Errorf("standard methods should be none")
	}

	if len(projectionMethods5) != 1 {
		t.Errorf("projection methods should be subsetted")
	}

	spec6 := &GenSpec{"", dummy, &GenTag{[]string{}}, nil}

	standardMethods6, projectionMethods6, err6 := determineMethods(spec6)

	if err6 != nil {
		t.Errorf("empty subsetted methods should be ok, instead got: '%v'", err6)
	}

	if len(standardMethods6) != 0 {
		t.Errorf("standard methods should be empty when the tag is empty")
	}

	if len(projectionMethods6) != 0 {
		t.Errorf("projection methods should be none")
	}
}

// +gen
type Thing1 int

type Thing2 Thing1

// +gen * methods:"Any,Where"
type Thing3 float64

// +gen projections:"int,Thing2"
type Thing4 struct{}

// +gen methods:"Count,GroupBy,Select,Aggregate" projections:"string,Thing4"
type Thing5 Thing4

// +gen projections:"float64,Thing1,Thing4,Thing7"
type Thing6 struct {
	Field int
}

// +gen
type Thing7 struct {
	Field func(int) // not comparable
}

// +gen
type Thing8 [7]Thing7

func TestStandardMethods(t *testing.T) {
	thing1, ok1 := typs["Thing1"]

	if !ok1 || thing1 == nil {
		t.Errorf("Thing1 should have been identified as a gen Type")
	}

	if len(thing1.Pointer) != 0 {
		t.Errorf("Thing1 should not generate pointers")
	}

	if len(thing1.StandardMethods) != len(StandardTemplates) {
		t.Errorf("Thing1 should have all standard methods")
	}

	if len(thing1.Projections) != 0 {
		t.Errorf("Thing1 should have no projections")
	}

	thing2, ok2 := typs["Thing2"]

	if ok2 || thing2 != nil {
		t.Errorf("Thing2 should not have been identified as a gen Type")
	}

	thing3 := typs["Thing3"]

	if thing3.Pointer != "*" {
		t.Errorf("Thing3 should generate pointers")
	}

	if len(thing3.StandardMethods) != 2 {
		t.Errorf("Thing3 should have subsetted Any and Where, but has: %v", thing3.StandardMethods)
	}

	if len(thing3.Projections) != 0 {
		t.Errorf("Thing3 should have no projections, but has: %v", thing3.Projections)
	}

	thing4 := typs["Thing4"]

	if len(thing4.Projections) != 2*len(ProjectionTemplates) {
		t.Errorf("Thing4 should have all projection methods for 2 typs, but has: %v", thing4.Projections)
	}

	thing5 := typs["Thing5"]

	if len(thing5.StandardMethods) != 1 {
		t.Errorf("Thing5 should have 1 subsetted standard method, but has: %v", thing5.StandardMethods)
	}

	if len(thing5.Projections) != 2*3 {
		t.Errorf("Thing4 should have 3 subsetted projection methods for 2 typs, but has: %v", thing5.Projections)
	}
}

func TestTypeEvaluation(t *testing.T) {
	thing6 := typs["Thing6"]
	methods6 := stringSliceToSet(thing6.StandardMethods)
	projections6 := projectionsToSet(thing6.Projections)

	if !methods6["Distinct"] {
		t.Errorf("Thing6 should have Distinct because it is comparable, but has: %v", thing6.StandardMethods)
	}

	if !projections6["AverageThing1"] {
		t.Errorf("Thing6 should have AverageThing1 because it is numeric, but has: %v", thing6.Projections)
	}

	if projections6["AverageThing4"] {
		t.Errorf("Thing6 should not have AverageThing4 because it is not numeric, but has: %v", thing6.Projections)
	}

	if !projections6["GroupByThing4"] {
		t.Errorf("Thing6 should have GroupByThing4 because it is comparable, but has: %v", thing6.Projections)
	}

	if projections6["GroupByThing7"] {
		t.Errorf("Thing6 should not have GroupByThing7 because it is not comparable, but has: %v", thing6.Projections)
	}

	thing7 := typs["Thing7"]
	methods7 := stringSliceToSet(thing7.StandardMethods)

	if methods7["Distinct"] {
		t.Errorf("Thing7 should not have Distinct because it is not comparable, but has: %v", thing7.StandardMethods)
	}

	thing8 := typs["Thing8"]
	methods8 := stringSliceToSet(thing8.StandardMethods)

	if methods8["Distinct"] {
		t.Errorf("Thing8 should not have Distinct because it is not comparable, but has: %v", thing8.StandardMethods)
	}
}

func TestSort(t *testing.T) {
	thing4 := typs["Thing4"]

	if !thing4.requiresSortSupport() {
		t.Errorf("Thing4 should require sort support. Methods: %v", thing4.StandardMethods)
	}

	thing5 := typs["Thing5"]

	if thing5.requiresSortSupport() {
		t.Errorf("Thing4 should not require sort support. Methods: %v", thing5.StandardMethods)
	}
}

func stringSliceToSet(s []string) map[string]bool {
	result := make(map[string]bool)
	for _, v := range s {
		result[v] = true
	}
	return result
}

func projectionsToSet(p []*Projection) map[string]bool {
	result := make(map[string]bool)
	for _, v := range p {
		result[v.MethodName()] = true
	}
	return result
}
