package pkg

import (
    "bufio"
    "bytes"
    "compress/gzip"
    "encoding/json"
)

func NDJSONGzip(objs []any) ([]byte, error) {
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    w := bufio.NewWriter(gz)
    for _, o := range objs {
        b, err := json.Marshal(o)
        if err != nil { return nil, err }
        if _, err := w.Write(b); err != nil { return nil, err }
        if err := w.WriteByte('\n'); err != nil { return nil, err }
    }
    if err := w.Flush(); err != nil { return nil, err }
    if err := gz.Close(); err != nil { return nil, err }
    return buf.Bytes(), nil
}
