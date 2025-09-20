package logsource

import (
    "bufio"
    "os"
    "time"
)

type FileTailer struct {
    Paths []string
    OnLine func(path string, line string)
    stop chan struct{}
}

func (t *FileTailer) Start() error {
    if t.stop != nil { return nil }
    t.stop = make(chan struct{})
    for _, p := range t.Paths {
        go t.tailFile(p, t.stop)
    }
    return nil
}

func (t *FileTailer) Stop() error { if t.stop != nil { close(t.stop) }; return nil }

func (t *FileTailer) tailFile(path string, stop <-chan struct{}) {
    // simple polling tail to end
    f, err := os.Open(path)
    if err != nil { return }
    defer f.Close()
    // seek to end
    if _, err := f.Seek(0, os.SEEK_END); err != nil { return }
    r := bufio.NewReader(f)
    for {
        select {
        case <-stop:
            return
        default:
        }
        line, err := r.ReadString('\n')
        if err != nil {
            time.Sleep(500 * time.Millisecond)
            continue
        }
        if t.OnLine != nil { t.OnLine(path, line) }
    }
}
