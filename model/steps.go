package model

type Step int

const (
	StopStep Step = iota
	StartLaunchStep
	AuthAccountStep
	GenerateCmdStep
	CompleteFilesStep
	ExecCmdStep
)
