package translate

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("translate", translate)
}

func translate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
