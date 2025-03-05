package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MoodyShoo/go-http-calculator/internal/models"
)

type Agent struct {
	config *Config
}

func New() *Agent {
	return &Agent{
		config: configFromEnv(),
	}
}

func (a *Agent) fetchTask() (*models.Task, error) {
	resp, err := http.Get("http://" + a.config.OrchestratorAddress + "/internal/task")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrator returned status: %d", resp.StatusCode)
	}

	var task models.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode task: %v", err)
	}

	return &task, nil
}

// TODO: context?
func (a *Agent) executeTask(task *models.Task) (float64, error) {
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

func (a *Agent) sendResult(taskId int, result float64, taskError error) error {
	resultData := models.TaskResult{
		Id:     taskId,
		Result: result,
	}

	if taskError != nil {
		resultData.Error = taskError.Error()
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	resp, err := http.Post("http://"+a.config.OrchestratorAddress+"/internal/task", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("orchestrator returned status: %d", resp.StatusCode)
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

func (a *Agent) Run() error {
	log.Printf("Agent running on: %s", a.config.Address)
	a.RunGoroutines(a.config.ComputingPower)
	return http.ListenAndServe(":"+a.config.Address, nil)
}
