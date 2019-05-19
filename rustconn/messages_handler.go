package rustconn

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/types"
)

// A Chat is for handling discord <-> rust chat
type messages struct {
	mChan   chan<- types.GameMessage
	timeout time.Duration
}

// newMessages initializes a chat handler and returns it
//
// in is the channel for server -> discord
func newMessages(in chan<- types.GameMessage) func(w http.ResponseWriter, r *http.Request) {
	m := messages{
		mChan:   in,
		timeout: 10 * time.Second,
	}

	return m.Handle
}

// Handle manages Rust <-> discord messages requests and logging
//
// HTTP POST requests are sent to the "in" chan
func (mh *messages) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	channel := vars["channel"]

	sc, err := getServerContext(r.Context())
	if err != nil {
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	mhLog := log.WithFields(logrus.Fields{"handler": "messages", "requestID": sc.requestUUID, "accountID": sc.account.ID.Hex(), "serverName": sc.server.Name})

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var message types.GameMessage

		err := decoder.Decode(&message)
		if err != nil {
			mhLog.WithError(err).Error("Invalid JSON")
			if err := handleError(w, types.RESTError{
				Error:      "Invalid request",
				StatusCode: http.StatusBadRequest,
			}); err != nil {
				mhLog.WithError(err).Error("http response failed to write")
			}
			return
		}

		message.Snowflake = sc.account.GuildSnowflake
		message.ChannelName = channel
		eChan := make(chan error)
		message.ErrorResponse = eChan

		mhLog.WithField("message", message).Trace("message")

		// sending message
		select {
		case mh.mChan <- message:
			break
		case <-time.After(mh.timeout):
			mhLog.Error("timed out sending message to channel")
			if err := handleError(w, types.RESTError{
				Error:      "internal error sending message to discord handler",
				StatusCode: http.StatusInternalServerError,
			}); err != nil {
				mhLog.WithError(err).Error("http response failed to write")
			}
			return
		}

		select {
		case err := <-eChan:
			if err != nil {
				var status int
				switch err.Error() {
				case "channel not found":
					status = http.StatusNotFound
				case "could not send to channel":
					status = http.StatusForbidden
				default:
					status = http.StatusInternalServerError
				}
				mhLog.WithError(err).Error("error from discord handler")
				if err := handleError(w, types.RESTError{
					Error:      err.Error(),
					StatusCode: status,
				}); err != nil {
					mhLog.WithError(err).Error("http response failed to write")
				}
			}
		case <-time.After(mh.timeout):
			mhLog.Error("timed out receiving discord response")
			if err := handleError(w, types.RESTError{
				Error:      "internal error receiving discord response",
				StatusCode: http.StatusInternalServerError,
			}); err != nil {
				mhLog.WithError(err).Error("http response failed to write")
			}
		}
	}
}
