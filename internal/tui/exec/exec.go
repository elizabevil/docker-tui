package exec

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type ExecStrategy int

const (
	StrategyAPI ExecStrategy = iota
	StrategyCLI
)

type ExecResult struct {
	Strategy ExecStrategy
	Session  runtimeapi.ExecSession
	CLIUsed  CLIType
	Error    error
}

type Executor interface {
	Exec(ctx context.Context, containerID, shell string, opts runtimeapi.ExecOptions) *ExecResult
}

type execExecutor struct {
	detector   CLIDetector
	executor   CLIExecutor
	cliEnabled bool
}

func NewExecutor(detector CLIDetector, executor CLIExecutor, cliEnabled bool) Executor {
	return &execExecutor{
		detector:   detector,
		executor:   executor,
		cliEnabled: cliEnabled,
	}
}

func (e *execExecutor) Exec(ctx context.Context, containerID, shell string, opts runtimeapi.ExecOptions) *ExecResult {
	if e.cliEnabled && e.detector != nil {
		result := e.detector.Detect(ctx, "", "")
		if result != nil && result.MatchError == nil {
			return &ExecResult{
				Strategy: StrategyCLI,
				CLIUsed:  result.CLIType,
			}
		}
	}

	return &ExecResult{
		Strategy: StrategyAPI,
	}
}

func DetectAndExecute(
	ctx context.Context,
	engine runtimeapi.Engine,
	containerID string,
	shell string,
	cliEnabled bool,
	detector CLIDetector,
	executor CLIExecutor,
	apiExec func() (runtimeapi.ExecSession, error),
) (runtimeapi.ExecSession, ExecStrategy, error) {

	if !cliEnabled || detector == nil {
		session, err := apiExec()
		return session, StrategyAPI, err
	}

	result := detector.Detect(ctx, "", "")
	if result != nil && result.MatchError == nil {
		err := executor.Execute(ctx, result.CLIPath, containerID, shell)
		if err == nil {
			return nil, StrategyCLI, nil
		}
	}

	session, err := apiExec()
	if err != nil {
		return nil, StrategyAPI, fmt.Errorf("CLI failed, API also failed: %w", err)
	}

	return session, StrategyAPI, nil
}
