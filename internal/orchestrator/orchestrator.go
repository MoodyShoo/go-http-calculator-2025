package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/MoodyShoo/go-http-calculator/internal/auth"
	"github.com/MoodyShoo/go-http-calculator/internal/database"
	"github.com/MoodyShoo/go-http-calculator/internal/middleware"
	"github.com/MoodyShoo/go-http-calculator/internal/models"
	pb "github.com/MoodyShoo/go-http-calculator/internal/proto"
	"github.com/MoodyShoo/go-http-calculator/internal/util"
	"github.com/MoodyShoo/go-http-calculator/pkg/calculation"
	"google.golang.org/grpc"
)

type Orchestrator struct {
	pb.OrchestratorServiceServer
	config     *Config
	db         *database.Database
	Ts         auth.TokenStore
	tasks      []*pb.Task
	nextTaskId int64
	mu         sync.Mutex
}

func New(db *database.Database) *Orchestrator {
	return &Orchestrator{
		config:     configFromEnv(),
		db:         db,
		Ts:         *auth.NewTokenStore(),
		tasks:      make([]*pb.Task, 0),
		nextTaskId: 1,
	}
}

// isNumber проверяет, является ли строка числом.
func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// operationTime возвращает время выполнения операции.
func (o *Orchestrator) operationTime(operation rune) int {
	switch operation {
	case '+':
		return o.config.TimeAdditionMs
	case '-':
		return o.config.TimeSubtractionMs
	case '*':
		return o.config.TimeMultiplicationsMs
	case '/':
		return o.config.TimeDivisionsMs
	default:
		return 0
	}
}

// createTasks создает задачи для выражения.
func (o *Orchestrator) createTasks(tokens []string, expressionId int64) ([]*pb.Task, error) {
	var tasks []*pb.Task
	var stack []string

	for _, token := range tokens {
		if isNumber(token) {
			stack = append(stack, token)
		} else if calculation.IsOperator(rune(token[0])) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("not enough operands for operator: %s", token)
			}

			arg2 := stack[len(stack)-1]
			arg1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			task := &pb.Task{
				Id:            int64(o.nextTaskId),
				ExpressionId:  int64(expressionId),
				Arg1:          arg1,
				Arg2:          arg2,
				Operation:     token,
				OperationTime: int64(o.operationTime(rune(token[0]))),
				Status:        string(models.StatusPending),
			}

			tasks = append(tasks, task)
			o.nextTaskId++
			stack = append(stack, fmt.Sprintf("task%d", task.Id))
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	return tasks, nil
}

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

func (o *Orchestrator) addTasks() error {
	expressions, err := o.db.ExpressionRepo.GetComputingAndPending()
	if err != nil {
		return err
	}

	for _, exp := range expressions {
		tokens, err := calculation.ShuntingYard(exp.Expr)
		if err != nil {
			return fmt.Errorf("failed to parse expression: %v", err)
		}

		tasks, err := o.createTasks(tokens, exp.Id)
		if err != nil {
			return fmt.Errorf("failed to create tasks: %v", err)
		}

		for _, task := range tasks {
			o.tasks = append(o.tasks, task)
			log.Printf("Added task id: %d; ExpressionId: %d; Arg1: %s; Arg2: %s; Operation: %s; OperationTime: %d;",
				task.Id, task.ExpressionId, task.Arg1, task.Arg2, task.Operation, task.OperationTime)
		}
	}

	return nil
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

func (o *Orchestrator) FetchTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Println("Invoked FetchTask: ", in)

	for i, task := range o.tasks {
		if task.Status == string(models.StatusPending) && !isTaskReference(task.Arg1) && !isTaskReference(task.Arg2) {
			task.Status = string(models.StatusComputing)
			o.tasks[i] = task

			expression, err := o.db.ExpressionRepo.GetExpressionByID(task.ExpressionId)
			if err != nil {
				return nil, err
			}

			if expression.Status != models.StatusDone {
				expression.Status = models.StatusComputing
				o.db.ExpressionRepo.UpdateExpression(task.ExpressionId, expression)
			}

			log.Println("sent task: ", task.Id)

			return &pb.TaskResponse{
				Task: &pb.Task{
					Id:            task.Id,
					ExpressionId:  task.ExpressionId,
					Arg1:          task.Arg1,
					Arg2:          task.Arg2,
					Operation:     task.Operation,
					OperationTime: task.OperationTime,
					Status:        task.Status,
					Result:        task.Result,
					Error:         task.Error,
				},
			}, nil
		}
	}

	log.Println("no tasks available")
	return nil, fmt.Errorf("no tasks available")
}

// SubmitTaskResult обрабатывает запрос на обновление результата задачи
func (o *Orchestrator) SendResult(ctx context.Context, in *pb.TaskResult) (*pb.SuccessResponse, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	var task *pb.Task
	var taskIndex int
	var found bool
	for i, t := range o.tasks {
		if t.Id == in.Id {
			task = t
			taskIndex = i
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("task not found")
	}

	// Обновляет статус задачи
	if in.Error != "" {
		task.Status = string(models.StatusError)
		task.Error = in.Error
	} else {
		task.Status = string(models.StatusDone)
		task.Result = in.Result
	}

	o.tasks[taskIndex] = task

	// Обновляет аргументы в других задачах, если они ссылаются на эту задачу
	for i, t := range o.tasks {
		if t.ExpressionId == task.ExpressionId && t.Status == string(models.StatusPending) {
			if strings.HasPrefix(t.Arg1, "task") && t.Arg1 == fmt.Sprintf("task%d", task.Id) {
				t.Arg1 = fmt.Sprintf("%f", task.Result)
			}
			if strings.HasPrefix(t.Arg2, "task") && t.Arg2 == fmt.Sprintf("task%d", task.Id) {
				t.Arg2 = fmt.Sprintf("%f", task.Result)
			}
			o.tasks[i] = t
		}
	}

	// Проверка, все ли задачи для этого выражения выполнены
	allTasksDone := true
	for _, t := range o.tasks {
		if t.ExpressionId == task.ExpressionId && t.Status != string(models.StatusDone) && t.Status != string(models.StatusError) {
			allTasksDone = false
			break
		}
	}

	if allTasksDone {
		expression, err := o.db.ExpressionRepo.GetExpressionByID(task.ExpressionId)
		if err != nil {
			return nil, err
		}
		expression.Result = task.Result
		if task.Status == string(models.StatusError) {
			expression.Status = models.StatusError
			expression.Error = task.Error
		} else {
			expression.Status = models.StatusDone
		}
		o.db.ExpressionRepo.UpdateExpression(task.ExpressionId, expression)

		// Удаляет все задачи, связанные с выполненным выражением
		var remainingTasks []*pb.Task
		for _, t := range o.tasks {
			if t.ExpressionId != expression.Id {
				remainingTasks = append(remainingTasks, t)
			}
		}
		o.tasks = remainingTasks
	}

	return &pb.SuccessResponse{Message: "Task result accepted."}, nil
}

// isTaskReference проверяет, является ли аргумент ссылкой на задачу
func isTaskReference(arg string) bool {
	return strings.HasPrefix(arg, "task")
}

// Хендлер регистрации
func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Логирование входящего запроса
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

	// Логирование входящего запроса
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

// RunServer запускает HTTP-сервер и gRPC сервер
func (o *Orchestrator) RunServer() error {
	http.HandleFunc(RegisterRoute, o.RegisterHandler)
	http.HandleFunc(LoginRoute, o.LoginHandler)
	http.HandleFunc(CalculateRoute, middleware.AuthMiddleware(&o.Ts, o.CalculateHandler))
	http.HandleFunc(ExpressionsRoute, middleware.AuthMiddleware(&o.Ts, o.ExpressionsHandler))
	http.HandleFunc(ExpressionIdRoute, middleware.AuthMiddleware(&o.Ts, o.ExpressionIdHandler))

	// горутина для gRPC сервера
	go func() {
		host := o.config.AddressGRPC
		port := o.config.PortGRPC
		addr := fmt.Sprintf("%s:%s", host, port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Println("error starting tcp listener:", err)
			os.Exit(1)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterOrchestratorServiceServer(grpcServer, o)

		log.Println("gRPC server started on", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Start gRPC server error: %v", err)
		}
	}()

	o.addTasks()

	log.Printf("HTTP server running on: %s", o.config.Address)
	return http.ListenAndServe(":"+o.config.Address, nil)
}
