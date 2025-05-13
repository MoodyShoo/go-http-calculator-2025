package orchestrator

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/MoodyShoo/go-http-calculator/internal/auth"
	"github.com/MoodyShoo/go-http-calculator/internal/database"
	"github.com/MoodyShoo/go-http-calculator/internal/middleware"
	"github.com/MoodyShoo/go-http-calculator/internal/models"
	pb "github.com/MoodyShoo/go-http-calculator/internal/proto"
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
