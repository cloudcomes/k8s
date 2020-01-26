/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
  "fmt"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/serializer"
  "k8s.io/apimachinery/pkg/version"
  "k8s.io/apiserver/pkg/server"

  "cloudcome.net/k8s-config/options"
  "github.com/spf13/cobra"
  restclient "k8s.io/client-go/rest"
)


var cfgFile string

// NewAPIServerCommand creates a *cobra.Command object with default parameters
func NewAPIServerCommand(option *options.ServerRunOptions, stopCh <-chan struct{}) *cobra.Command {
  s := option
  cmd := &cobra.Command{
    Use: "k8s-config",
    Long: `The demo for Options and Config.`,
    RunE: func(cmd *cobra.Command, args []string) error {
      fmt.Println("Hello CLI")

      fmt.Print(s.SecureServing.BindPort)

      if err := RunServer(s,stopCh); err != nil {
        return err
      }
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

func fakeVersion() version.Info {
  return version.Info{
    Major:        "42",
    Minor:        "42",
    GitVersion:   "42",
    GitCommit:    "34973274ccef6ab4dfaaf86599792fa9c3fe4689",
    GitTreeState: "Dirty",
  }
}

func Config() server.Config {
  scheme := runtime.NewScheme()
  codecs := serializer.NewCodecFactory(scheme)

  config := server.NewConfig(codecs)

  return *config
}

// Config returns config for the api server given ServerRunOptions
func buildGenericConfig(option *options.ServerRunOptions) server.Config {
  // launch server
  opt := option
  config := Config()

  v := fakeVersion()
  config.Version = &v

  config.EnableIndex = true
  config.LoopbackClientConfig = &restclient.Config{}

  secureOptions := &options.SecureServingOptions{
    BindAddress: opt.SecureServing.BindAddress,
    BindPort:    opt.SecureServing.BindPort,
    Listener: nil,
  }

  // use a random free port
  //ln, err := net.Listen("tcp", "127.0.0.1:0")
  //if err != nil {
  //  fmt.Errorf("failed to listen on 127.0.0.1:0")
  //}

  //secureOptions.Listener = ln
  // get port
  //secureOptions.BindPort = ln.Addr().(*net.TCPAddr).Port

  if err := secureOptions.ApplyTo(&config.SecureServing); err != nil {
    fmt.Errorf("failed applying the SecureServingOptions: %v", err)
  }

  return config
}

// RunServer starts a new Server given ServerOptions
func RunServer(option *options.ServerRunOptions,stopCh <-chan struct{}) error {

  config := buildGenericConfig(option)

  s, err := config.Complete(nil).New("test", server.NewEmptyDelegate())
  if err != nil {
    fmt.Errorf("failed creating the server: %v", err)
  }

  // add poststart hook to know when the server is up.

  startedCh := make(chan struct{})
  s.AddPostStartHookOrDie("test-notifier", func(context server.PostStartHookContext) error {
    close(startedCh)
    return nil
  })

  return s.PrepareRun().Run(stopCh)
}



