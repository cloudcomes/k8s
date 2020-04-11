package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRESTMapperVersionAndKindForResource() {
	testGroup := "test.group"
	testVersion := "test"
	testGroupVersion := schema.GroupVersion{Group: testGroup, Version: testVersion}

	testCases := []struct {
		Resource               schema.GroupVersionResource
		GroupVersionToRegister schema.GroupVersion
		ExpectedGVK            schema.GroupVersionKind
		Err                    bool
	}{
		{Resource: schema.GroupVersionResource{Resource: "internalobjec"}, Err: true},
		{Resource: schema.GroupVersionResource{Resource: "internalObjec"}, Err: true},

		{Resource: schema.GroupVersionResource{Resource: "internalobject"}, ExpectedGVK: testGroupVersion.WithKind("InternalObject")},
		{Resource: schema.GroupVersionResource{Resource: "internalobjects"}, ExpectedGVK: testGroupVersion.WithKind("InternalObject")},
	}
	for i, testCase := range testCases {
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{testGroupVersion})
		if len(testCase.ExpectedGVK.Kind) != 0 {
			mapper.Add(testCase.ExpectedGVK, meta.RESTScopeNamespace)
		}
		actualGVK, err := mapper.KindFor(testCase.Resource)

		hasErr := err != nil
		if hasErr != testCase.Err {
			fmt.Println("%d: unexpected error behavior %t: %v", i, testCase.Err, err)
			continue
		}
		if err != nil {
			continue
		}

		if actualGVK != testCase.ExpectedGVK {
			fmt.Println("%d: unexpected version and kind: e=%s a=%s", i, testCase.ExpectedGVK, actualGVK)
		}
	}
}

func main() {
	TestRESTMapperVersionAndKindForResource()
}
