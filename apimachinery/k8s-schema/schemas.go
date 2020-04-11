package main

import (
	runtimetesting "cloudcome.net/k8s-schema/testing"
	"fmt"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"reflect"
	"strings"
)

// Returns a new Scheme set up with the test objects.
func GetTestScheme() *runtime.Scheme {
	internalGV := schema.GroupVersion{Version: runtime.APIVersionInternal}
	externalGV := schema.GroupVersion{Version: "v1"}
	alternateExternalGV := schema.GroupVersion{Group: "custom", Version: "v1"}
	alternateInternalGV := schema.GroupVersion{Group: "custom", Version: runtime.APIVersionInternal}
	//differentExternalGV := schema.GroupVersion{Group: "other", Version: "v2"}

	s := runtime.NewScheme()
	// Ordinarily, we wouldn't add TestType2, but because this is a test and
	// both types are from the same package, we need to get it into the system
	// so that converter will match it with ExternalType2.
	s.AddKnownTypes(internalGV, &runtimetesting.TestType1{}, &runtimetesting.TestType2{}, &runtimetesting.ExternalInternalSame{})
	//s.AddKnownTypes(externalGV, &runtimetesting.ExternalInternalSame{})
	s.AddKnownTypeWithName(externalGV.WithKind("TestType1"), &runtimetesting.ExternalTestType1{})
	//s.AddKnownTypeWithName(externalGV.WithKind("TestType2"), &runtimetesting.ExternalTestType2{})
	s.AddKnownTypeWithName(internalGV.WithKind("TestType3"), &runtimetesting.TestType1{})
	//s.AddKnownTypeWithName(externalGV.WithKind("TestType3"), &runtimetesting.ExternalTestType1{})
	//s.AddKnownTypeWithName(externalGV.WithKind("TestType4"), &runtimetesting.ExternalTestType1{})
	s.AddKnownTypeWithName(alternateInternalGV.WithKind("TestType3"), &runtimetesting.TestType1{})
	s.AddKnownTypeWithName(alternateExternalGV.WithKind("TestType3"), &runtimetesting.ExternalTestType1{})
	//s.AddKnownTypeWithName(alternateExternalGV.WithKind("TestType5"), &runtimetesting.ExternalTestType1{})
	//s.AddKnownTypeWithName(differentExternalGV.WithKind("TestType1"), &runtimetesting.ExternalTestType1{})
	s.AddUnversionedTypes(externalGV, &runtimetesting.UnversionedType{})

	utilruntime.Must(s.AddConversionFuncs(func(in *runtimetesting.TestType1, out *runtimetesting.ExternalTestType1, s conversion.Scope) error {
		out.A = in.A
		return nil
	}))
	return s
}


func TestConvert() {
	testCases := []struct {
		scheme *runtime.Scheme
		in     runtime.Object
		into   runtime.Object
		gv     runtime.GroupVersioner
		out    runtime.Object
		errFn  func(error) bool
	}{
		// converts from internal to unstructured, given a target version
		{
			scheme: GetTestScheme(),
			in:     &runtimetesting.TestType1{A: "test"},
			into:   &runtimetesting.Unstructured{},
			out: &runtimetesting.Unstructured{Object: map[string]interface{}{
				"myVersionKey": "custom/v1",
				"myKindKey":    "TestType3",
				"A":            "test",
			}},
			gv: schema.GroupVersion{Group: "custom", Version: "v1"},
		},
	}
	for i,test := range testCases {
		    fmt.Println(i)
		    fmt.Printf("in Object is :%#v\n", test.in)
			err := test.scheme.Convert(test.in, test.into, test.gv)
			switch {
			case test.errFn != nil:
				if !test.errFn(err) {
					fmt.Println("unexpected error: %v", err)
				}
				return
			case err != nil:
				fmt.Println("unexpected error: %v", err)
				return
			}

		    fmt.Printf("out Object is:%#v", test.out)
			if !reflect.DeepEqual(test.into, test.out) {
				fmt.Println("unexpected out: %s", err)
			}

	}
}


func TestConvertToVersion(){

	internalGV := schema.GroupVersion{Group: "test.group", Version: runtime.APIVersionInternal}
	internalGVK := internalGV.WithKind("Simple")
	externalGV := schema.GroupVersion{Group: "test.group", Version: "testExternal"}
	externalGVK := externalGV.WithKind("Simple")

	s := runtime.NewScheme()
	s.AddKnownTypeWithName(internalGVK, &runtimetesting.InternalSimple{})
	s.AddKnownTypeWithName(externalGVK, &runtimetesting.ExternalSimple{})


	for gvk, apiType := range s.AllKnownTypes() {

		fmt.Println("Checking %s", gvk.Version)
		// we do not care about internal objects or lists // TODO make sure this is always true
		if gvk.Version == runtime.APIVersionInternal || strings.HasSuffix(apiType.Name(), "List") {
			continue
		}

		intGV := gvk.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()

		tt := &runtimetesting.ExternalSimple{TestString: "I'm not a pointer object"}
		other, err := s.ConvertToVersion(tt, intGV)
		fmt.Printf("out Object v1 :%#v\n", other)

		if err != nil {
			log.Printf("unexpected error creating %s: %v", gvk, err)
			continue
		}

		//intGV := gvk.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
		// ConvertToVersion attempts to convert an input object to its matching Kind in another
		// version within this scheme.
		obj, err := s.New(gvk)
		fmt.Println(reflect.TypeOf(obj))
		//output: *testing.ExternalSimple
		fmt.Printf("in Object is :%#v\n", obj)

		intObj, err := s.ConvertToVersion(obj, intGV)
		if err != nil {
			log.Printf("unexpected error converting %s to internal: %v", gvk, err)
			continue
		}
		fmt.Println(reflect.TypeOf(intObj))
        //output: *testing.InternalSimple
		fmt.Printf("out Object is :%#v\n", intObj)
	}
}


func TestEncode() {
	internalGV := schema.GroupVersion{Group: "test.group", Version: runtime.APIVersionInternal}
	internalGVK := internalGV.WithKind("Simple")
	externalGV := schema.GroupVersion{Group: "test.group", Version: "testExternal"}
	externalGVK := externalGV.WithKind("Simple")

	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(internalGVK, &runtimetesting.InternalSimple{})
	scheme.AddKnownTypeWithName(externalGVK, &runtimetesting.ExternalSimple{})

	// CodecFactory provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(externalGV)

	test := &runtimetesting.InternalSimple{
		TestString: "I'm the same",
	}
	obj := runtime.Object(test)
	data, err := runtime.Encode(codec, obj)

	fmt.Println(string(data))

	obj2, gvk, err2 := codec.Decode(data, nil, nil)
	if err != nil || err2 != nil {
		fmt.Println("Failure: '%v' '%v'", err, err2)
	}

	intsimple:= obj2.(*runtimetesting.InternalSimple)

	if intsimple.TestString != "I'm the same" {
		fmt.Println("unexpected decoded object: %#v", intsimple)
	}
	fmt.Println(string(intsimple.TestString))

	if !reflect.DeepEqual(obj2, test) {
		fmt.Println("Expected:\n %#v,\n Got:\n %#v", test, obj2)
	}
	if *gvk != externalGVK {
		fmt.Println("unexpected gvk returned by decode: %#v", *gvk)
	}
}


// AddUnversionedTypes registers the provided types as "unversioned", which means that they follow special rules.
// Whenever an object of this type is serialized, it is serialized with the provided group version and is not
// converted. Thus unversioned objects are expected to remain backwards compatible forever, as if they were in an
// API group and version that would never be updated.

func TestUnversionedTypes() {
	internalGV := schema.GroupVersion{Group: "test.group", Version: runtime.APIVersionInternal}
	internalGVK := internalGV.WithKind("Simple")
	externalGV := schema.GroupVersion{Group: "test.group", Version: "testExternal"}
	externalGVK := externalGV.WithKind("Simple")
	otherGV := schema.GroupVersion{Group: "group", Version: "other"}

	scheme := runtime.NewScheme()
	scheme.AddUnversionedTypes(externalGV, &runtimetesting.InternalSimple{})
	scheme.AddKnownTypeWithName(internalGVK, &runtimetesting.InternalSimple{})

	scheme.AddKnownTypeWithName(externalGVK, &runtimetesting.ExternalSimple{})
	scheme.AddKnownTypeWithName(otherGV.WithKind("Simple"), &runtimetesting.ExternalSimple{})


	codec := serializer.NewCodecFactory(scheme).LegacyCodec(externalGV)

	if unv, ok := scheme.IsUnversioned(&runtimetesting.InternalSimple{}); !unv || !ok {
		fmt.Println("type not unversioned and in scheme: %t %t", unv, ok)
	}

	kinds, _, err := scheme.ObjectKinds(&runtimetesting.InternalSimple{})
	if err != nil {
		fmt.Println(err)
	}
	kind := kinds[0]
	if kind != externalGV.WithKind("InternalSimple") {
		fmt.Println("unexpected: %#v", kind)
	}

	test := &runtimetesting.InternalSimple{
		TestString: "I'm the same",
	}
	obj := runtime.Object(test)
	data, err := runtime.Encode(codec, obj)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("obj is :%#v\n", obj)
	fmt.Println("Json data is :%#s\n",string(data))

	obj2, gvk, err := codec.Decode(data, nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("obj2 is :%#v\n", obj2)



	if _, ok := obj2.(*runtimetesting.InternalSimple); !ok {
		fmt.Println("Got wrong type")
	}
	if !reflect.DeepEqual(obj2, test) {
		fmt.Println("Expected:\n %#v,\n Got:\n %#v", test, obj2)
	}
	// object is serialized as an unversioned object (in the group and version it was defined in)
	if *gvk != externalGV.WithKind("InternalSimple") {
		fmt.Println("unexpected gvk returned by decode: %#v", *gvk)
	}

	// when serialized to a different group, the object is kept in its preferred name
	codec = serializer.NewCodecFactory(scheme).LegacyCodec(otherGV)
	data, err = runtime.Encode(codec, obj)
	fmt.Printf("obj is :%#v\n", obj)

	fmt.Println("Json data is :%#s\n",string(data))

	if err != nil {
		fmt.Println(err)
	}
	if string(data) != `{"apiVersion":"test.group/testExternal","kind":"InternalSimple","testString":"I'm the same"}`+"\n" {
		fmt.Println("unexpected data: %s", data)
	}
}



func TestEncodev() {
	internalGV := schema.GroupVersion{Group: "test.group", Version: runtime.APIVersionInternal}
	internalGVK := internalGV.WithKind("Simple")
	externalGV := schema.GroupVersion{Group: "test.group", Version: "testExternal"}
	externalGVK := externalGV.WithKind("Simple")

	otherGV := schema.GroupVersion{Group: "group", Version: "v1"}
	otherGV1 := schema.GroupVersion{Group: "group", Version: "v1"}
	otherGV2 := schema.GroupVersion{Group: "group", Version: "v2"}

	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(internalGVK, &runtimetesting.InternalSimple{})
	scheme.AddKnownTypeWithName(externalGVK, &runtimetesting.ExternalSimple{})

	scheme.AddKnownTypeWithName(otherGV.WithKind("Simple"), &runtimetesting.ExternalSimple{})
	scheme.AddKnownTypeWithName(otherGV1.WithKind("Inter1"), &runtimetesting.InternalSimple{})
	scheme.AddKnownTypeWithName(otherGV2.WithKind("Inter2"), &runtimetesting.InternalSimple{})


	//scheme.AddUnversionedTypes(externalGV, &runtimetesting.InternalSimple{})

	// CodecFactory provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(externalGV)

	test := &runtimetesting.ExternalSimple{
		TestString: "I'm the same",
	}
	obj := runtime.Object(test)
	data, _ := runtime.Encode(codec, obj)
	fmt.Println(string(data))


	codec1 := serializer.NewCodecFactory(scheme).LegacyCodec(otherGV1)
	test1 := &runtimetesting.InternalSimple{
		TestString: "I'm the same",
	}
	obj1 := runtime.Object(test1)
	data1, _ := runtime.Encode(codec1, obj1)
	fmt.Println(string(data1))

}

func main() {
	//TestConvert()
	//TestConvertToVersion()
	//TestEncode()
	//TestUnversionedTypes()
	//TestEncodev()


}
