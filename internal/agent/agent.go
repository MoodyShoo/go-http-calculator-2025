package agent

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/MoodyShoo/go-http-calculator/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	config *Config
	client pb.OrchestratorServiceClient
}

func New() *Agent {
	conf := configFromEnv()

	conn, err := grpc.NewClient(conf.OrchestratorGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewOrchestratorServiceClient(conn)

	return &Agent{
		config: conf,
		client: client,
	}
}

func (a *Agent) fetchTask() (*pb.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	task, err := a.client.FetchTask(ctx, &pb.TaskRequest{})

	if err != nil {
		return nil, fmt.Errorf("could not get task: %v", err)
	}

	return task.Task, nil
}

func (a *Agent) executeTask(task *pb.Task) (float64, error) {
	timer := time.NewTimer(time.Duration(task.OperationTime) * time.Millisecond)
	defer timer.Stop()

	<-timer.C

	arg1, err := parseArg(task.Arg1)
	if err != nil {
		return 0, fmt.Errorf("invalid Arg1: %v", err)
	}

	arg2, err := parseArg(task.Arg2)
	if err != nil {
		return 0, fmt.Errorf("invalid Arg2: %v", err)
	}

	switch task.Operation {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return arg1 / arg2, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", task.Operation)
	}
}

// parseArg преобразует строку в число, если это возможно
func parseArg(arg string) (float64, error) {
	if strings.HasPrefix(arg, "task") {
		return 0, fmt.Errorf("agent cannot handle task references: %s", arg)
	}
	return strconv.ParseFloat(arg, 64)
}

func (a *Agent) sendResult(taskId int64, result float64, taskError error) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	taskResult := &pb.TaskResult{
		Id:     taskId,
		Result: result,
	}

	if taskError != nil {
		taskResult.Error = taskError.Error()
	}

	_, err := a.client.SendResult(ctx, taskResult)

	if err != nil {
		return fmt.Errorf("could not send result: %v", err)
	}
	return nil
}

func (a *Agent) RunGoroutines(num int) {
	for i := 0; i < num; i++ {
		go func(workerID int) {
			for {
				// Каждый воркер агента обращается каждые две секунды (Сделал для демонстрации и дебага)
				time.Sleep(2 * time.Second)

				// Получаем задачу
				task, err := a.fetchTask()
				if err != nil {
					log.Printf("Worker %d: Error fetching task: %v", workerID, err)
					continue
				}

				log.Printf("Worker %d: Received task %d (Expression %d): %s %s %s",
					workerID, task.Id, task.ExpressionId, task.Arg1, task.Operation, task.Arg2)

				// Выполняем задачу
				result, err := a.executeTask(task)
				if err != nil {
					log.Printf("Worker %d: Error executing task %d: %v", workerID, task.Id, err)
				}

				// Отправляем результат
				if err := a.sendResult(task.Id, result, err); err != nil {
					log.Printf("Worker %d: Error sending result for task %d: %v", workerID, task.Id, err)
				} else {
					log.Printf("Worker %d: Result for task %d sent successfully.", workerID, task.Id)
				}
			}
		}(i + 1)
	}
}

func (a *Agent) Run() {
	var wt sync.WaitGroup
	wt.Add(a.config.ComputingPower)
	log.Printf("Agent listens: %s", a.config.OrchestratorGRPC)
	a.RunGoroutines(a.config.ComputingPower)

	wt.Wait()
}
