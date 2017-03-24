package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/feedhenry/negotiator/pkg/deploy"
	"github.com/feedhenry/negotiator/pkg/log"
	"github.com/feedhenry/negotiator/pkg/status"
	"github.com/gorilla/mux"
)

type StatusRetriever interface {
	Get(key string) (*deploy.ConfigurationStatus, error)
}

type LastActionHandler struct {
	statusRetriever StatusRetriever
	logger          log.Logger
}

func NewLastActionHandler(statusRet StatusRetriever, logger log.Logger) LastActionHandler {
	return LastActionHandler{
		statusRetriever: statusRet,
		logger:          logger,
	}
}
func (lah LastActionHandler) LastAction(rw http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	instance := params["instance_id"]
	planID := params["plan_id"]      // not currently used
	operation := params["operation"] // provision, update , deprovision
	if planID == "" {
		planID = "noplan"
	}
	statusKey := strings.Join([]string{instance, planID, operation}, ":")
	status, err := lah.statusRetriever.Get(statusKey)
	if err != nil {
		lah.handleError(err, "failed to retrieve status", rw)
		return
	}
	encoder := json.NewEncoder(rw)
	if err := encoder.Encode(status); err != nil {
		lah.handleError(err, "failed encoding response ", rw)
		return
	}

}

func (lah LastActionHandler) handleError(err error, msg string, rw http.ResponseWriter) {
	switch err.(type) {
	case *status.ErrNotExist:
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(msg + err.Error()))
		return
	case *json.SyntaxError, deploy.ErrInvalid:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(msg + err.Error()))
		return
	}
	lah.logger.Error(fmt.Sprintf(" unexpected error getting last operation. context: %s \n %+v", msg, err))
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write([]byte(msg + err.Error()))
}
