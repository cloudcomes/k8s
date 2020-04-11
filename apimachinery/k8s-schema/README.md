# Schema

The API Server keeps all known Kubernetes object kinds in a Go type registry called Scheme. In this registry, each version of kinds are defined along with how they can be converted, how new objects can be created, and how objects can be encoded and decoded to JSON or protobuf.

###  Clone Project.
```
$ https://github.com/cloudcomes/k8s.git
```

### Go to k8s-options folder 
```
$ cd k8s-schema
```

### Run cli Application

```
$ go run schemas.go
```

### Notes
这个实例演示了runtime.Scheme里序列化、反序列化的核心方法New()的代码。        
通过查找gkvToType里匹配的类型，以反射方法生成一个空的数据对象：

```
// New returns a new API object of the given version and name, or an error if it hasn't
// been registered. The version and kind fields must be specified.
func (s *Scheme) New(kind schema.GroupVersionKind) (Object, error) {
    if t, exists := s.gvkToType[kind]; exists {
        return reflect.New(t).Interface().(Object), nil
    }

    if t, exists := s.unversionedKinds[kind.Kind]; exists {
        return reflect.New(t).Interface().(Object), nil
    }
    return nil, NewNotRegisteredErrForKind(kind)
}
```


由于这个方法的实现，也让Scheme实现了ObjectCreater接口。
