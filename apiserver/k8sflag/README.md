# Cobra & pflag

###  Initialize the modules.
```
$ go mod init cloudcome.net
```

### Initialize the cli scaffolding 
```
$ cobra init k8sflag --pkg-name cloudcome.net/k8sflag
```

###  Add flag in root.go   


```
Run: func(cmd *cobra.Command, args []string) {

  fmt.Println("Hello CLI")
  fmt.Println(BindPort)
},
```


```
func init() {
  cobra.OnInitialize(initConfig)

  // Here you will define your flags and configuration settings.
  // Cobra supports persistent flags, which, if defined here,
  // will be global for your application.

  fss := rootCmd.Flags()
  // Create the set of flags for your kubect-plugin
  fs := pflag.NewFlagSet("demo", pflag.ExitOnError)

  //pflag.CommandLine = fs
  fss.IntVar(&BindPort, "secure-port", 80, "Port")

  fss.AddFlagSet(fs)
}
```


### Build cli Application

```
$ go build .
```

### Run cli Application

```
$ ./k8sflag --secure-port 443
```

> Hello CLI      
> 443