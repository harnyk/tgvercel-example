package handler

import (
	"net/http"

	"github.com/harnyk/tgvercel"
)

var tgv = tgvercel.New(tgvercel.DefaultOptions())

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleSetup(w, r)
}
