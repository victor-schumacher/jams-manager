package rider

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BorsaTeam/jams-manager/server"
	"github.com/BorsaTeam/jams-manager/server/database/repository"
	"github.com/BorsaTeam/jams-manager/server/error/errors"
)

var riders = server.Riders{}

type Manager struct {
	riderRepository repository.Rider
}

func NewHandler(rider repository.Rider) Manager {
	return Manager{riderRepository: rider}
}

func (m Manager) Handle() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			m.processGet(w, r)
		case http.MethodPost:
			m.processPost(w, r)
		case http.MethodDelete:
			processDelete(r)
		case http.MethodPut:
			processPut(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func (m Manager) processGet(w http.ResponseWriter, r *http.Request) {
	if id := id(r.URL.Path); id != "" {
		rider, err := m.findOne(id)
		if err != nil {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(errors.Unknown)
			return
		}
		if rider.Id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rider)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findAll())
}

func (m Manager) findOne(id string) (server.Rider, error) {

	riderEntity, err := m.riderRepository.FindOne(id)
	if err != nil {
		return server.Rider{}, err
	}
	sponsorsString := strings.Split(riderEntity.Sponsors, ",")

	rider := server.Rider{
		Id:               riderEntity.Id,
		Name:             riderEntity.Name,
		Age:              riderEntity.Age,
		Gender:           riderEntity.Gender,
		City:             riderEntity.City,
		Cpf:              riderEntity.Cpf,
		PaidSubscription: riderEntity.PaidSubscription,
		Sponsors:         sponsorsString,
		CategoryId:       riderEntity.CategoryId,
	}

	return rider, nil
}

func findAll() server.Riders {
	return riders
}

func (m Manager) processPost(w http.ResponseWriter, r *http.Request) {
	rider := server.Rider{}

	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&rider)
	if err != nil {
		w.Write([]byte("Error while processing data"))
		return
	}

	riderId, err := m.createRider(rider)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error while processing data RIDER"))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Write([]byte(riderId))
}

func (m Manager) createRider(r server.Rider) (string, error) {
	sponsorsString := strings.Join(r.Sponsors, ",")

	riderEntity := repository.RiderEntity{
		Id:               r.Id,
		Name:             r.Name,
		Age:              r.Age,
		Gender:           r.Gender,
		City:             r.City,
		Cpf:              r.Cpf,
		PaidSubscription: r.PaidSubscription,
		Sponsors:         sponsorsString,
		CategoryId:       r.CategoryId,
		CreateAt:         time.Now(),
	}

	riderId, err := m.riderRepository.Save(riderEntity)
	if err != nil {
		return "", err
	}

	return riderId, nil
}

func processPut(w http.ResponseWriter, r *http.Request) {
	rider := server.Rider{}

	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&rider)
	if err != nil {
		w.Write([]byte("Error while processing data"))
	}

	id := id(r.URL.Path)
	for i := range riders {
		if riders[i].Id == id {
			riders[i] = rider
		}
	}
}

func id(path string) string {
	p := strings.Split(path, "/")
	if len(p) > 1 {
		return p[2]
	}
	return ""
}

func processDelete(r *http.Request) {
	id := id(r.URL.Path)

	for i := range riders {
		if riders[i].Id == id {
			riders = append(riders[:i], riders[i+1:]...)
			break
		}
	}
}
