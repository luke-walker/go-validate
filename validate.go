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

func (v Validator) ValidateData(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var data map[string]any
        switch contentType := r.Header.Get("Content-Type"); contentType {
        case "application/x-www-form-urlencoded":
            if err := r.ParseForm(); err != nil {
                http.Error(w, "Invalid request payload (URL-encoded form expected)", http.StatusBadRequest)
                return
            }
            data = make(map[string]any)
            for key, values := range r.PostForm {
                if len(values) == 1 {
                    data[key] = values[0]
                } else {
                    data[key] = values
                }
            }
            break
        case "application/json":
            var bodyBuf bytes.Buffer
            body := io.TeeReader(r.Body, &bodyBuf)
            r.Body = io.NopCloser(&bodyBuf)
            if err := json.NewDecoder(body).Decode(&data); err != nil {
                http.Error(w, "Invalid request payload (JSON expected)", http.StatusBadRequest)
                return
            }
            break
        default:
            http.Error(w, fmt.Sprintf("Unsupported request payload (%s)", contentType), http.StatusBadRequest)
            return
        }

        for field, cons := range v.Fields {
            if _, ok := data[field]; !ok {
                if required, ok := cons["required"].(bool); ok && required {
                    http.Error(w, fmt.Sprintf("Missing required field '%s'", field), http.StatusBadRequest)
                    return
                }
                continue
            }
        }

        next.ServeHTTP(w, r)
    })
}
