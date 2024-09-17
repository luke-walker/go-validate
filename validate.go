package govalidate

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type FieldsMap map[string]map[string]any

type Validator struct {
    // field:
    //   required: bool
    //   type: type (WIP)
    Fields FieldsMap
}

func (v Validator) ValidateJSON(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var bodyBuf bytes.Buffer
        body := io.TeeReader(r.Body, &bodyBuf)
        r.Body = io.NopCloser(&bodyBuf)

        var data map[string]any
        if err := json.NewDecoder(body).Decode(&data); err != nil {
            http.Error(w, "Invalid request payload (JSON expected)", http.StatusBadRequest)
            return
        }

        for field, cons := range v.Fields {
            _, ok := data[field]
            if !ok {
                required, ok := cons["required"].(bool)
                if !ok || !required {
                    continue
                }

                http.Error(w, fmt.Sprintf("Missing required field '%s'", field), http.StatusBadRequest)
                return
            }
        }

        next.ServeHTTP(w, r)
    })
}
