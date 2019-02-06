package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func (hub *Hub) wsHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &WsClient{
		Conn: conn,
	}

	hub.RLock()
	for _, r := range hub.rotators {
		ev := Event{
			Name:        AddRotator,
			RotatorName: r.Name(),
		}
		if err := c.write(ev); err != nil {
			fmt.Println(err)
		}
	}
	hub.RUnlock()

	hub.addWsClient(c)
}

func (hub *Hub) rotatorsHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	rotators := hub.serializeRotators()

	if err := json.NewEncoder(w).Encode(rotators); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to encode rotator msg"))
	}
}

func (hub *Hub) rotatorHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	if err := json.NewEncoder(w).Encode(r.Serialize()); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to encode rotatorData to json"))
	}
}

func (hub *Hub) azimuthHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	switch req.Method {
	case "GET":

		rs := rotator.AzimuthGet{
			HasAzimuth: r.HasAzimuth(),
			Azimuth:    r.Azimuth(),
			Preset:     r.AzPreset(),
		}

		if err := json.NewEncoder(w).Encode(rs); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to encode rotatorData to json"))
		}

	case "PUT":
		azPUT := rotator.AzimuthPut{}
		dec := json.NewDecoder(req.Body)

		if err := dec.Decode(&azPUT); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid json"))
			return
		}

		if azPUT.Azimuth == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
			return
		}

		if !r.HasAzimuth() {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(("rotator does not support azimuth")))
			return
		}

		err := r.SetAzimuth(*azPUT.Azimuth)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("unable to set azimuth to %v: %s", *azPUT.Azimuth, err)))
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (hub *Hub) elevationHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	switch req.Method {
	case "GET":

		rs := rotator.ElevationGet{
			HasElevation: r.HasElevation(),
			Elevation:    r.Elevation(),
			Preset:       r.ElPreset(),
		}

		if err := json.NewEncoder(w).Encode(rs); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to encode rotatorData to json"))
		}

	case "PUT":
		elPUT := rotator.ElevationPut{}
		dec := json.NewDecoder(req.Body)

		if err := dec.Decode(&elPUT); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid json"))
			return
		}

		if elPUT.Elevation == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
			return
		}

		if !r.HasElevation() {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(("rotator does not support elevation")))
			return
		}

		err := r.SetElevation(*elPUT.Elevation)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("unable to set elevation to %v: %s", *elPUT.Elevation, err)))
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (hub *Hub) stopAzimuthHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	if !r.HasAzimuth() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(("rotator does not support azimuth")))
		return
	}

	err := r.StopAzimuth()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to stop rotator: %v", err.Error())))
		return
	}
}

func (hub *Hub) stopElevationHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	if !r.HasElevation() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(("rotator does not support elevation")))
		return
	}

	err := r.StopElevation()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to stop rotator: %v", err.Error())))
		return
	}
}

func (hub *Hub) stopHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(req)
	rName := vars["rotator"]

	r, ok := hub.Rotator(rName)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to find rotator"))
		return
	}

	err := r.Stop()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to stop rotator: %v", err.Error())))
		log.Println(err)
		return
	}
}

func (hub *Hub) serializeRotators() rotator.Objects {

	hub.RLock()
	defer hub.RUnlock()

	rs := rotator.Objects{}

	for _, r := range hub.rotators {
		sr := r.Serialize()
		rs[sr.Name] = sr
	}

	return rs
}
