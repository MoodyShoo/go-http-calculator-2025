package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/MoodyShoo/go-http-calculator/internal/middleware"
	"github.com/MoodyShoo/go-http-calculator/internal/models"
	"github.com/MoodyShoo/go-http-calculator/internal/util"
)

// handleCalculateRequest обрабатывает запрос на вычисление выражения.
func (o *Orchestrator) handleCalculateRequest(req models.Request, userId int64) (int64, error) {
	exp := models.Expression{
		Expr:   req.Expression,
		Status: models.StatusPending,
		UserID: userId,
	}

	id, err := o.db.ExpressionRepo.InsertExpression(exp)
	if err != nil {
		return 0, fmt.Errorf("failed to insert expression: %v", err)
	}

	err = o.addTasks()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// CalculateHandler обрабатывает HTTP-запрос на вычисление выражения
func (o *Orchestrator) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("CalculateHandler: started")
	defer log.Printf("CalculateHandler: finished")

	o.mu.Lock()
	defer o.mu.Unlock()

	var req models.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CalculateHandler: failed to decode request body: %v", err)
		util.SendError(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	if req.Expression == "" {
		util.SendError(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("CalculateHandler: processing expression: %s", req.Expression)

	userId, ok := middleware.GetUserID(r)
	if !ok {
		util.SendError(w, "user ID not found in context", http.StatusUnauthorized)
		return
	}

	expressionId, err := o.handleCalculateRequest(req, userId)
	if err != nil {
		log.Printf("CalculateHandler: %v", err)
		util.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.SendResponse(w, &models.AcceptedResponse{Id: expressionId}, http.StatusAccepted)
}

// ExpressionsHandler возвращает список всех выражений
func (o *Orchestrator) ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ExpressionsHandler: started")
	defer log.Printf("ExpressionsHandler: finished")

	userId, ok := middleware.GetUserID(r)
	if !ok {
		util.SendError(w, "user ID not found in context", http.StatusUnauthorized)
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	response, err := o.db.ExpressionRepo.GetExpressionsByUser(userId)
	if err != nil {
		util.SendError(w, err.Error(), http.StatusInternalServerError)
	}

	if response != nil {
		sort.Slice(response, func(i, j int) bool {
			return response[i].Id < response[j].Id
		})
	} else {
		response = make([]models.Expression, 0)
	}

	util.SendResponse(w, &models.ExpressionsResponse{Expressions: response}, http.StatusOK)
}

// ExpressionIdHandler возвращает выражение по его ID
func (o *Orchestrator) ExpressionIdHandler(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("ExpressionIdHandler: started")
	defer log.Printf("ExpressionIdHandler: finished")

	idStr := strings.TrimPrefix(r.URL.Path, ExpressionIdRoute)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		util.SendError(w, "invalid ID", http.StatusBadRequest)
		return
	}

	userId, ok := middleware.GetUserID(r)
	if !ok {
		util.SendError(w, "user ID not found in context", http.StatusUnauthorized)
		return
	}

	expression, exists := o.db.ExpressionRepo.GetExpressionByIDByUser(id, userId)
	if exists != nil {
		util.SendError(w, "expression not found", http.StatusNotFound)
		return
	}

	util.SendResponse(w, &expression, http.StatusOK)
}

// Хендлер регистрации
func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("RegisterHandler: received %s request", r.Method)

	if r.Method != http.MethodPost {
		log.Printf("RegisterHandler: invalid method %s", r.Method)
		util.SendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("RegisterHandler: failed to decode request body: %v", err)
		util.SendError(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	if req.Login == "" || req.Password == "" {
		util.SendError(w, "login or password can't be empty", http.StatusUnauthorized)
		return
	}

	log.Printf("RegisterHandler: registering user with login %s", req.Login)

	err := o.db.UserRepo.AddUser(req.Login, req.Password)
	if err != nil {
		log.Printf("RegisterHandler: failed to add user %s: %v", req.Login, err)
		util.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("RegisterHandler: successfully registered user %s", req.Login)

	w.WriteHeader(http.StatusOK)
}

// Хендлер логина
func (o *Orchestrator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("LoginHandler: received %s request", r.Method)

	if r.Method != http.MethodPost {
		log.Printf("LoginHandler: invalid method %s", r.Method)
		util.SendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("LoginHandler: failed to decode request body: %v", err)
		util.SendError(w, "unprocessable entity", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("LoginHandler: attempting to login user %s", req.Login)

	user, err := o.db.UserRepo.GetUser(req.Login, req.Password)
	if err != nil {
		log.Printf("LoginHandler: failed to authenticate user %s: %v", req.Login, err)
		util.SendError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("LoginHandler: user %s authenticated successfully", req.Login)

	token, err := o.Ts.AddToken(user.Id)
	if err != nil {
		log.Printf("LoginHandler: failed to create token for user %s: %v", req.Login, err)
		util.SendError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("LoginHandler: token created for user %s", req.Login)

	util.SendResponse(w, &models.AuthResponse{Token: token}, http.StatusOK)
}
