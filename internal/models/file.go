package models

import "time"

type FileInfo struct {
    FileName     string    `json:"fileName"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"lastModified"`
}