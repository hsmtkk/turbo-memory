package save

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("save", save)
}

func save(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
