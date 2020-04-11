package funs

import (
	"fmt"
	"k8s.io/apimachinery/pkg/conversion"
	"reflect"
)

// nameFunc returns the name of the type that we wish to use to determine when two types attempt
// a conversion. Defaults to the go name of the type if the type is not registered.


func TestConverter_byteSlice() {
	c := conversion.NewConverter(conversion.DefaultNameFunc)
	src := []byte{1, 2, 3}
	dest := []byte{}
	fmt.Print("Before convert\n")
	fmt.Println(src)
	fmt.Println(dest)
	err := c.Convert(&src, &dest, 0, nil)
	if err != nil {
		fmt.Println("expected no error")
	}
	if e, a := src, dest; !reflect.DeepEqual(e, a) {
		fmt.Println("expected %#v, got %#v", e, a)
	}
	fmt.Print("After convert\n")
	fmt.Println(src)
	fmt.Println(dest)
}

func TestConverter_DefaultConvert() {
	type A struct {
		Foo string
		Baz int
	}
	type B struct {
		Bar string
		Baz int
	}

    var nameFunc = func(t reflect.Type) string { return "MyType" }
	c := conversion.NewConverter(nameFunc)

	// Ensure conversion funcs can call DefaultConvert to get default behavior,
	// then fixup remaining fields manually
	err := c.RegisterConversionFunc(func(in *A, out *B, s conversion.Scope) error {
		fmt.Print("Register Conversion Fun is called\n")
		if err := s.DefaultConvert(in, out, conversion.IgnoreMissingFields); err != nil {
			return err
		}
		out.Bar = in.Foo
		return nil
	})
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}

	x := A{"hello, intrepid test reader!", 3}
	y := B{}

	fmt.Print("Before convert\n")

	fmt.Printf("x is :%#v\n", x)
	fmt.Printf("y is :%#v\n", y)
	err = c.Convert(&x, &y, 0, nil)
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}
	if e, a := x.Foo, y.Bar; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}
	if e, a := x.Baz, y.Baz; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}

	fmt.Print("After convert\n")
	fmt.Printf("x is :%#v\n", x)
	fmt.Printf("y is :%#v\n", y)


}


func TestConverter_CallsRegisteredFunctions() {
	type A struct {
		Foo string
		Baz int
	}
	type B struct {
		Bar string
		Baz int
	}
	type C struct{}
	c := conversion.NewConverter(conversion.DefaultNameFunc)

	err := c.RegisterConversionFunc(func(in *A, out *B, s conversion.Scope) error {
		fmt.Print("Register Conversion Fun is called\n")
		out.Bar = in.Foo
		return s.Convert(&in.Baz, &out.Baz, 0)
	})
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}
	err = c.RegisterConversionFunc(func(in *B, out *A, s conversion.Scope) error {
		fmt.Print("Register Conversion Fun is called\n")
		out.Foo = in.Bar
		return s.Convert(&in.Baz, &out.Baz, 0)
	})
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}

	x := A{"hello, intrepid test reader!", 3}
	y := B{}

	fmt.Print("Before convert\n")
	fmt.Printf("x is :%#v\n", x)
	fmt.Printf("y is :%#v\n", y)

	err = c.Convert(&x, &y, 0, nil)
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}
	if e, a := x.Foo, y.Bar; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}
	if e, a := x.Baz, y.Baz; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}

	fmt.Print("After convert\n")
	fmt.Printf("x is :%#v\n", x)
	fmt.Printf("y is :%#v\n", y)

	//##############################

	z := B{"all your test are belong to us", 42}
	w := A{}

	fmt.Print("Before convert\n")
	fmt.Printf("z is :%#v\n", z)
	fmt.Printf("w is :%#v\n", w)
	err = c.Convert(&z, &w, 0, nil)
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}
	if e, a := z.Bar, w.Foo; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}
	if e, a := z.Baz, w.Baz; e != a {
		fmt.Println("expected %v, got %v", e, a)
	}
	fmt.Print("After convert\n")
	fmt.Printf("z is :%#v\n", z)
	fmt.Printf("w is :%#v\n", w)

	err = c.RegisterConversionFunc(func(in *A, out *C, s conversion.Scope) error {
		fmt.Print("Register Conversion Fun is called:in *A,out *C\n")
		return fmt.Errorf("C can't store an A, silly")
	})
	if err != nil {
		fmt.Println("unexpected error %v", err)
	}

	fmt.Print("Before convert\n")
	fmt.Printf("A is :%#v\n", A{})
	fmt.Printf("B is :%#v\n", B{})
	err = c.Convert(&A{}, &C{}, 0, nil)
	if err == nil {
		fmt.Println("unexpected non-error")
	}

	fmt.Print("After convert\n")
	fmt.Printf("A is :%#v\n", A{})
	fmt.Printf("B is :%#v\n", B{})
}

func TestConverter_IgnoredConversion() {
	type A struct{}
	type B struct{}

	count := 0
	c := conversion.NewConverter(conversion.DefaultNameFunc)
	if err := c.RegisterConversionFunc(func(in *A, out *B, s conversion.Scope) error {
		count++
		return nil
	}); err != nil {
		fmt.Println("unexpected error %v", err)
	}
	if err := c.RegisterIgnoredConversion(&A{}, &B{}); err != nil {
		fmt.Println(err)
	}
	a := A{}
	b := B{}

	fmt.Print("Before convert\n")
	fmt.Printf("a is :%#v\n", a)
	fmt.Printf("b is :%#v\n", b)

	if err := c.Convert(&a, &b, 0, nil); err != nil {
		fmt.Println("%v", err)
	}

	fmt.Print("After convert\n")
	fmt.Printf("a is :%#v\n", a)
	fmt.Printf("b is :%#v\n", b)

	if count != 0 {
		fmt.Println("unexpected number of conversion invocations")
	}


}

func TestConverter_meta() {
	type Foo struct{ A string }
	type Bar struct{ A string }
	c := conversion.NewConverter(conversion.DefaultNameFunc)

	checks := 0
	err := c.RegisterConversionFunc(
		func(in *Foo, out *Bar, s conversion.Scope) error {
			if s.Meta() == nil {
				fmt.Println("Meta did not get passed!")
			}
			fmt.Print("Register Conversion Fun is called\n")
			checks++
			//out.A = in.A
			s.Convert(&in.A, &out.A, 0)
			return nil
		},
	)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
	}
	err = c.RegisterConversionFunc(
		func(in *string, out *string, s conversion.Scope) error {
			if s.Meta() == nil {
				fmt.Println("Meta did not get passed a second time!")
			}
			fmt.Print("Register Conversion Fun is called\n")
			checks++
			*out=*in
			return nil
		},
	)
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
	}


	x := Foo{"hello, intrepid test reader!"}
	y := Bar{}

	fmt.Print("Before convert\n")
	fmt.Printf("Foo is :%#v\n", x)
	fmt.Printf("Bar is :%#v\n", y)

	err = c.Convert(&x, &y, 0, &conversion.Meta{})
	if err != nil {
		fmt.Println("Unexpected error: %v", err)
	}
	fmt.Print("After convert\n")
	fmt.Printf("Foo is :%#v\n", x)
	fmt.Printf("Bar is :%#v\n", y)


	if checks != 2 {
		fmt.Println("Registered functions did not get called.")
	}
}


