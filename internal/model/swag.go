//go:build ignore
// +build ignore

package model

type WebApiError struct {
	Err string `json:"error,omitempty" example:"problem explanation"`
}

type WebApiSuccess struct {
	Data int64 `json:"data,omitempty" example:"5"`
}

type WebApiPageSuccess struct {
	Meta Meta        `json:"meta,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
