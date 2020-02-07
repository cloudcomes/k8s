package main
import (
	"github.com/emicklei/go-restful"
	"k8s.io/apiserver/pkg/server/mux"
	"log"
	"net/http"
)


// HandlerChainBuilderFn is used to wrap the GoRestfulContainer handler using the provided handler chain.
// It is normally used to apply filtering like authentication and authorization
type HandlerChainBuilderFn func(apiHandler http.Handler) http.Handler

type director struct {
	name               string
	goRestfulContainer *restful.Container
	nonGoRestfulMux    *mux.PathRecorderMux
}



func middlewareOne(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing middlewareOne")
		next.ServeHTTP(w, r)
		log.Println("Executing middlewareOne again")
	})
}

func middlewareTwo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing middlewareTwo")
		if r.URL.Path != "/" {
			return
		}
		next.ServeHTTP(w, r)
		log.Println("Executing middlewareTwo again")
	})
}

func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing finalHandler")
	w.Write([]byte("OK"))
}

func main() {
	finalHandler := http.HandlerFunc(final)

	http.Handle("/", middlewareOne(middlewareTwo(finalHandler)))
	http.ListenAndServe(":3000", nil)
}

//2020/02/03 15:45:55 Executing middlewareOne
//2020/02/03 15:45:55 Executing middlewareTwo
//2020/02/03 15:45:55 Executing finalHandler
//2020/02/03 15:45:55 Executing middlewareTwo again
//2020/02/03 15:45:55 Executing middlewareOne again

