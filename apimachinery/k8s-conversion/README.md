# Schema

 Package conversion provides go object versioning.

Specifically, conversion provides a way for you to define multiple versions
of the same object. You may write functions which implement conversion logic,
but for the fields which did not change, copying is automated. This makes it
easy to modify the structures you use in memory without affecting the format
you store on disk or respond to in your external API calls.

Converter完成不同Object之间的转换,转换函数通过注册的方式添加到Converter中,或者使用缺省的转换函数；                            
K8SConverter用来把一个版本的K8S API object转换成其他版本的K8S API object。


###  Clone Project.
```
$ https://github.com/cloudcomes/k8s.git
```

### Go to k8s-options folder 
```
$ cd k8s-conversions
```

### Run cli Application

```
$ go run convert.go
```

