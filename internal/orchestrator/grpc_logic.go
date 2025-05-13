package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MoodyShoo/go-http-calculator/internal/models"
	pb "github.com/MoodyShoo/go-http-calculator/internal/proto"
)

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
