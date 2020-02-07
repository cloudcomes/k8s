# Cobra & pflag & Options

Demonstration of how to use the pflag, Cobra and Options library to build a functional k8s.io/apiserver API server.

###  Clone Project.
```
$ https://github.com/cloudcomes/k8s.git
```

### Go to k8s-options folder 
```
$ cd k8s-handler
```

### Build cli Application

```
$ go build .
```

### Run cli Application

```
$ ./k8s-handler 
```
### Access service
Make a request to http://localhost:8080/apis,You should get output similar to this in IE:    
`APIServerHandler ServeHTTP---->WithPanicRecovery middleware---->Director ServeHTTP---->GoRestfulContainer hanlder `
     
Make a request to http://localhost:8080/,You should get output similar to this in IE:    
`APIServerHandler ServeHTTP---->WithPanicRecovery middleware---->Director ServeHTTP---->nonGoRestfulMux ServeHTTP`
