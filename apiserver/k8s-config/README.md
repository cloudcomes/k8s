# Cobra & pflag & Options & Config
Demonstration of how to use the pflag, Cobra, Options and Config library to build a functional k8s.io/apiserver API server.

The config created here contains runnable data structures, i.e. configs are runtime objects, in contrast to the options which correspond to flags.
###  Clone Project.
```
$ https://github.com/cloudcomes/k8s.git
```

### Go to k8s-options folder 
```
$ cd k8s-config
```

### Build cli Application

```
$ go build .
```

### Run cli Application

```
$ ./k8s-options --bind-address 127.0.0.1 --secure-port 4436
```

>W0126 17:58:51.151996   19155 authorization.go:47] Authorization is disabled           
>W0126 17:58:51.152112   19155 authentication.go:92] Authentication is disabled  
I0126 17:58:51.153689   19155 secure_serving.go:178] Serving securely on 127.0.0.1:4436

