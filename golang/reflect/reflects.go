package golang



import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"reflect"
)

type Pod struct {
	Name string
}

type Node struct {
	Name string
}

func (pod Pod) GetObjectKind() schema.ObjectKind {
 return  nil
}

func (node Node) GetObjectKind() schema.ObjectKind {
 return  nil
}

func (pod Pod) DeepCopyObject() Object {
	return  nil
}

func (node Node) DeepCopyObject() Object {
	return  nil
}


// Object interface must be supported by all API types registered with Scheme. Since objects in a scheme are
// expected to be serialized to the wire, the interface an Object must provide to the Scheme allows
// serializers to set the kind, version, and group the object is represented as. An Object may choose
// to return a no-op ObjectKindAccessor in cases where it is not expected to be serialized.
type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}


func main() {
	podType := reflect.TypeOf(Pod{})

	nodeType := reflect.TypeOf(Node{})

	types := []reflect.Type{
		podType, nodeType,
	}

	for _, t := range types {
		log.Printf("type is %s", t.String())

		v := reflect.New(t).Interface()
		log.Printf("reflect new type is %T", v)

		v1 := reflect.New(t).Interface().(Object)
		log.Printf("reflect new type Object is %T", v1)

		var v2 interface{}
		if t.Kind() != reflect.Ptr {
			v2 = reflect.New(t).Elem().Interface()
		}
		log.Printf("reflect new elem type is %T", v2)
	}
}

//clouds-MacBook-Pro:golang cloud$ go run reflects.go
//2020/02/17 11:51:52 type is main.Pod
//2020/02/17 11:51:52 reflect new type is *main.Pod
//2020/02/17 11:51:52 reflect new type Object is *main.Pod
//2020/02/17 11:51:52 reflect new elem type is main.Pod

//2020/02/17 11:51:52 type is main.Node
//2020/02/17 11:51:52 reflect new type is *main.Node
//2020/02/17 11:51:52 reflect new type Object is *main.Node
//2020/02/17 11:51:52 reflect new elem type is main.Node