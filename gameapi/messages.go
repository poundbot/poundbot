package gameapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/types"
)

// A Chat is for handling discord <-> rust chat
type messages struct {
	mChan        chan<- types.GameMessage
	channelsChan chan<- types.ServerChannelsRequest
	timeout      time.Duration
}

// initMessages initializes a chat handler and returns it
//
// mChan is the channel for server -> discord
func initMessages(mChan chan<- types.GameMessage, cChan chan<- types.ServerChannelsRequest, api *mux.Router) {
	m := messages{
		mChan:   mChan,
		timeout: 10 * time.Second,
	}

	api.HandleFunc("/messages", m.rootHandler).
		Methods(http.MethodPost)

	api.HandleFunc("/messages/{channel}", m.channelHandler).
		Methods(http.MethodPost)
}

func (mh *messages) rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
}

// Handle manages Rust <-> discord messages requests and logging
//
// HTTP POST requests are sent to the "in" chan
func (mh *messages) channelHandler(w http.ResponseWriter, r *http.Request) {
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

	decoder := json.NewDecoder(r.Body)
	var message types.GameMessage

	err = decoder.Decode(&message)
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
