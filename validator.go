package validator

import (
	"strings"
	"errors"
	"reflect"
	"fmt"
	"context"
)

var (
	ErrorValidatorNotFound = errors.New("validator is not found")
	ErrorType              = errors.New("conv type error")
)

type CustomValidField struct {
	required bool
	name     string
	value    interface{}
	ctx      context.Context
	vfs      []CustomValidFunc
}

func (c *CustomValidField) Name() (string) {
	return c.name
}

func (c *CustomValidField) Interface() (interface{}, error) {
	return c.value, nil
}

func (c *CustomValidField) Int64() (int64, error) {
	if c.value == nil {
		return 0, nil
	} else {
		v, ok := c.value.(int64)
		if ok {
			return v, nil
		} else {
			v1, ok := c.value.(*int64)
			if ok {
				return *v1, nil
			}
		}
	}
	return 0, ErrorType
}

func (c *CustomValidField) String() (string, error) {
	if c.value == nil {
		return "", nil
	} else {
		v, ok := c.value.(string)
		if ok {
			return v, nil
		} else {
			v1, ok := c.value.(*string)
			if ok {
				return *v1, nil
			}
		}
	}
	return "", ErrorType
}

type CustomValidFunc func(CustomValidField) (error)

type validator struct {
	// TODO	不支持并行校验
	cvf map[string]CustomValidFunc
	err error
}

func (v *validator) RegisterValidation(name string, fn CustomValidFunc) {
	v.cvf[name] = fn
}

func (v *validator) Validate(ctx context.Context, u interface{}) {
	t := reflect.TypeOf(u)
	value := reflect.ValueOf(u)
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).CanInterface() { // 判断是否为可导出字段
			vStr := t.Field(i).Tag.Get("validate")
			if len(strings.TrimSpace(vStr)) != 0 {
				cvf := CustomValidField{}
				validators := strings.Split(vStr, ",")
				for i := 0; i < len(validators); i++ {
					validators[i] = strings.TrimSpace(validators[i])
					if validators[i] == "required" {
						cvf.required = true
					} else {
						f, ok := v.cvf[validators[i]]
						if !ok {
							v.err = ErrorValidatorNotFound
						}
						cvf.vfs = append(cvf.vfs, f)
					}
				}
				cvf.name = fmt.Sprintf("%s.%s", t.Name(), t.Field(i).Name)
				cvf.value = value.Field(i).Interface()
				cvf.ctx = ctx

				if !(t.Field(i).Type.Kind() == reflect.Ptr && value.Field(i).IsNil()) {
					for _, f := range cvf.vfs {
						v.err = f(cvf)
						if v.err != nil {
							return
						}
					}
				}
			}
			ctx = context.WithValue(ctx, t.Field(i).Name, value.Field(i).Interface())
		}
	}
}

func New() *validator {
	v := validator{
		cvf: make(map[string]CustomValidFunc),
	}
	return &v
}
