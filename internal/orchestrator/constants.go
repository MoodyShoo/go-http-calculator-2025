package orchestrator

const (
	RegisterRoute     = "/api/v1/register"
	LoginRoute        = "/api/v1/login"
	CalculateRoute    = "/api/v1/calculate"
	ExpressionsRoute  = "/api/v1/expressions"
	ExpressionIdRoute = "/api/v1/expressions/"
	TaskRoute         = "/internal/task"

	ContentType     = "Content-Type"
	ApplicationJson = "application/json"

	PortEnv                  = "PORT"
	GRPCAddressEnv           = "GRPC_ADDRESS"
	GRPCPortEnv              = "GRPC_PORT"
	TimeAdditionMsEnv        = "TIME_ADDITION_MS"
	TimeSubtractionMsEnv     = "TIME_SUBTRACTION_MS"
	TimeMultiplicationsMsEnv = "TIME_MULTIPLICATIONS_MS"
	TimeDivisionsMsEnv       = "TIME_DIVISIONS_MS"
)
