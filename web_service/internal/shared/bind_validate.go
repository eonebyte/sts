package shared

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// BindAndValidate hanya decode & validate, return error biasa
func BindAndValidate(r *http.Request, v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.New("bind target must be pointer")
	}

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		log.Printf(
			"[DECODE] invalid request body | path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)
		return errors.New("invalid request body")
	}

	if err := validate.Struct(v); err != nil {
		// kita tidak mengubah tipe, cukup kembalikan error sederhana
		return errors.New("validation failed")
	}

	return nil
}
