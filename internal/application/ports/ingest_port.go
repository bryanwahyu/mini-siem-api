package ports

type Tailer interface {
    Start() error
    Stop() error
}
