package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"faceit/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type daoClient interface {
	Get(ctx context.Context, id string) (*model.User, error)
	Insert(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userId string) error
	Filter(ctx context.Context, conditions []*model.FilterCondition) ([]*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
}

type msgClient interface {
	Publish(ctx context.Context, msg *model.Message) error
}

type Handler struct {
	db  daoClient
	msg msgClient
}

func NewHandler(db daoClient, msg msgClient) *Handler {
	return &Handler{
		db:  db,
		msg: msg,
	}
}

func (h *Handler) GetUser(r *http.Request) (int, interface{}, error) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	log.WithField("id", id).Info("retrieve user")
	user, err := h.db.Get(ctx, id)
	if err != nil {
		// Here an internal error looks like a missing entry, extra code can distinguish
		log.WithField("id", id).Error(fmt.Sprintf("unable to retrieve id. err: %v", err))
		err = fmt.Errorf("unable to find user: %s", id)
		return http.StatusNotFound, nil, err
	}
	log.WithField("user", user).Info("retrieved user")
	return http.StatusOK, user, nil
}

func (h *Handler) AddUser(r *http.Request) (int, interface{}, error) {
	ctx := r.Context()
	user := &model.User{
		Id: uuid.New().String(),
	}

	log.Info("unmarshal request")
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		log.Error("unable to unmarshal request")
		return http.StatusBadRequest, nil, err
	}

	log.WithField("user", user).Info("insert user")
	err = h.db.Insert(ctx, user)
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to store user")
		return http.StatusInternalServerError, nil, errors.New("unable to store user")
	}
	log.WithField("user", user).Info("publish message")
	err = h.msg.Publish(ctx, model.NewMessage(user.Id, model.UserAdd))
	if err != nil {
		// User was still created, for the sake of the exercise not going to implement any kind of rollback
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to publish message, user created")
		return http.StatusCreated, user, nil
	}
	return http.StatusCreated, user, nil
}

func (h *Handler) RemoveUser(r *http.Request) (int, interface{}, error) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	log.WithField("id", id).Info("check for user")
	user, err := h.db.Get(ctx, id)
	if err != nil {
		log.WithField("id", id).Error(fmt.Sprintf("unable to retrieve id. err: %v", err))
		err = fmt.Errorf("unable to find user: %s", id)
		return http.StatusNotFound, nil, err
	}

	log.WithField("id", id).Info("delete user")
	err = h.db.Delete(ctx, id)
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to delete user")
		return http.StatusInternalServerError, nil, fmt.Errorf("unable to remove user: %s", id)
	}
	log.WithField("id", id).Info("publish message")
	err = h.msg.Publish(ctx, model.NewMessage(user.Id, model.UserDelete))
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to publish message, user deleted")
		return http.StatusNoContent, nil, nil
	}
	return http.StatusNoContent, nil, nil
}

func (h *Handler) UpdateUser(r *http.Request) (int, interface{}, error) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	log.WithField("id", id).Info("check for user")
	user, err := h.db.Get(ctx, id)
	if err != nil {
		log.WithField("id", id).Error(fmt.Sprintf("unable to retrieve id. err: %v", err))
		return http.StatusNotFound, nil, fmt.Errorf("unable to find user: %s", id)
	}

	log.Info("unmarshal request")
	update := &model.User{}
	err = json.NewDecoder(r.Body).Decode(update)
	if err != nil {
		log.Error("unable to unmarshal request")
		return http.StatusBadRequest, nil, err
	}

	log.WithField("user", user).Info("insert updated user")
	update.Id = user.Id
	err = h.db.Insert(ctx, update)
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to store user")
		return http.StatusInternalServerError, nil, fmt.Errorf("unable to update user: %s", id)
	}
	log.WithField("id", id).Info("publish message")
	err = h.msg.Publish(ctx, model.NewMessage(user.Id, model.UserDelete))
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err,
		}).Error("unable to publish message, user deleted")
		return http.StatusOK, user, nil
	}
	return http.StatusOK, update, nil
}

func (h *Handler) FilterUsers(r *http.Request) (int, interface{}, error) {
	ctx := r.Context()

	log.Info("determine filter params")
	conditions := []*model.FilterCondition{}
	if len(r.URL.Query()) == 0 {
		log.Info("no query params, return all")
		return h.GetAllUsers(ctx)
	}
	log.Info("prepare filter conditions")
	for query, value := range r.URL.Query() {
		condition, ok := prepareFilter(query, value)
		if !ok {
			msg := fmt.Sprintf("malformed filter query %s: %s", query, value)
			log.Error(msg)
			return http.StatusBadRequest, nil, errors.New(msg)
		}
		conditions = append(conditions, condition)
	}

	log.Info("filter users")
	results, err := h.db.Filter(ctx, conditions)
	if err != nil {
		log.WithFields(log.Fields{
			"results": results,
			"error":   err,
		}).Error("unable to filter users")
		return http.StatusInternalServerError, nil, errors.New("unable to search for users")
	}
	if results == nil {
		log.Info("no results found for filters")
		return http.StatusOK, "no results found", nil
	}
	log.WithField("results", results).Info("filtered users")
	response := &model.FilterResponse{
		Results: results,
		Count:   len(results),
	}
	return http.StatusOK, response, nil
}

// perpareFilter is a slight convenience function, and also allows for extra conditions / handling of alternative types
func prepareFilter(query string, value []string) (*model.FilterCondition, bool) {
	if len(value) == 0 {
		return nil, false
	}
	switch query {
	case "country", "nickname", "surname", "forename", "email", "password": // Do we really want to filter on password?
		return &model.FilterCondition{
			Query: query,
			Value: value[0], // assuming one value per query param
		}, true
	default:
		return nil, false
	}
}

func (h *Handler) GetAllUsers(ctx context.Context) (int, interface{}, error) {
	log.Info("retrieve all users")
	results, err := h.db.GetAll(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"results": results,
			"error":   err,
		}).Error("unable to retrieve users")
		return http.StatusInternalServerError, nil, errors.New("unable to retrieve users")
	}
	response := &model.FilterResponse{
		Results: results,
		Count:   len(results),
	}
	log.WithField("results", results).Info("filtered users")
	return http.StatusOK, response, nil
}
