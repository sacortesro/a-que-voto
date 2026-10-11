package controllers

import (
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/utils"
)

func HandleKeepAlive(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJson(w, http.StatusOK, struct{}{})
}
