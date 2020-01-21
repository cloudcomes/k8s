# Cobra & pflag & Options

###  Initialize the modules.
```
$ go mod init cloudcome.net
```

### Initialize the cli scaffolding 
```
$ cobra init k8s-options --pkg-name cloudcome.net/k8s-options
```

###  Add Options
   >serving.go: HTTPS configuration    
   >options.go: contains flags and options for initializing an apiserver    
   >sectioned.go: define named flag sets      

###  Add NewAPIServerCommand() in root.go   


```
func NewAPIServerCommand() *cobra.Command {
  s := options.NewServerRunOptions()
  cmd := &cobra.Command{
    Use: "k8s-options",
    Long: `The Kubernetes API server validates and configures data
for the api objects which include pods, services, replicationcontrollers, and
others. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
    RunE: func(cmd *cobra.Command, args []string) error {
      fmt.Println("Hello CLI")
      //fmt.Println()

      return nil
    },
  }

  fs := cmd.Flags()
  namedFlagSets := s.Flags()
  for _, f := range namedFlagSets.FlagSets {
    fs.AddFlagSet(f)
  }

  return cmd
}

```


### Build cli Application

```
$ go build .
```

### Run cli Application

```
$ ./k8s-options -h
```

