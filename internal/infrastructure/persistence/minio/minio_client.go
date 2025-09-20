package minio

import (
    "context"
    "net/http"
    "time"

    minio "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
    MC *minio.Client
}

func New(endpoint string, useSSL bool, accessKey, secretKey string, region string, timeoutSec int, maxRetries int) (*Client, error) {
    if timeoutSec <= 0 { timeoutSec = 10 }
    tr := defaultTransport(timeoutSec)
    c, err := minio.New(endpoint, &minio.Options{
        Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure:    useSSL,
        Region:    region,
        Transport: tr,
    })
    if err != nil { return nil, err }
    c.SetAppInfo("server-analyst", "v0")
    return &Client{MC: c}, nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucket, region string) error {
    exists, err := c.MC.BucketExists(ctx, bucket)
    if err != nil { return err }
    if !exists {
        return c.MC.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
    }
    return nil
}

// helper to create a transport with timeout
func defaultTransport(timeoutSec int) *http.Transport {
    t := &http.Transport{Proxy: http.ProxyFromEnvironment}
    to := time.Duration(timeoutSec) * time.Second
    t.ResponseHeaderTimeout = to
    t.IdleConnTimeout = to
    t.TLSHandshakeTimeout = to
    return t
}
