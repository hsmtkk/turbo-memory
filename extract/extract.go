package extract

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("extract", extract)
}

func extract(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
