package pipeline

type PipelineComponent interface {
	EventLoop() error
}
