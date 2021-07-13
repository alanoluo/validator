package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
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

func (c *CustomValidField) Name() string {
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

type CustomValidFunc func(CustomValidField) error

type Validator struct {
	// TODO	不支持并行校验
	cvf map[string]CustomValidFunc
}

func (v *Validator) RegisterValidation(name string, fn CustomValidFunc) *Validator {
	v.cvf[name] = fn
	return v
}

func (v *Validator) Validate(ctx context.Context, u interface{}) error {
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
							return ErrorValidatorNotFound
						}
						cvf.vfs = append(cvf.vfs, f)
					}
				}
				cvf.name = fmt.Sprintf("%s.%s", t.Name(), t.Field(i).Name)
				cvf.value = value.Field(i).Interface()
				cvf.ctx = ctx

				if !(t.Field(i).Type.Kind() == reflect.Ptr && value.Field(i).IsNil()) {
					for _, f := range cvf.vfs {
						err := f(cvf)
						if err != nil {
							return err
						}
					}
				} else if cvf.required {
					return errors.New(fmt.Sprintf("MissParam.%s", t.Field(i).Name))
				}
			}
			fmt.Printf("Name:%s, %v\n", t.Field(i).Name, value.Field(i).Interface())
			ctx = context.WithValue(ctx, t.Field(i).Name, value.Field(i).Interface())
			fmt.Printf("value:%v\n", ctx)
		}
	}
	return nil
}

func New() *Validator {
	v := Validator{
		cvf: make(map[string]CustomValidFunc),
	}
	return &v
}






// ListEventBusesRequest 获取事件集列表入参
type ListEventBusesRequest struct {
	Limit   *int64  `json:"Limit" validate:"ValidateLimitRange"`
	OrderBy *string `json:"OrderBy" validate:"ValidateEventBusOrderByRange"`
}

// ValidateEventBusOrderByRange 检验<OrderBy>参数
func ValidateEventBusOrderByRange(field CustomValidField) error {
	data := []string{"AddTime", "ModTime"}
	for _, v := range data {
		d, _ := field.String()
		if v == d {
			return nil
		}
	}
	return errors.New("InvalidParameterValueOrderBy")
}

// ValidateLimitRange 检验<Limit>参数
func ValidateLimitRange(field CustomValidField) error {
	v, _ := field.Int64()
	if v < 0 || v > 20 {
		return errors.New("InvalidParameterValueLimit")
	}
	return nil
}

func main() {
	v := New()
	v.RegisterValidation("ValidateEventBusOrderByRange", ValidateEventBusOrderByRange)
	v.RegisterValidation("ValidateLimitRange", ValidateLimitRange)

	d := int64(20)
	m := "ASC"
	data := ListEventBusesRequest{
		Limit: &d,
		OrderBy : &m,
	}

	ctx := context.Background()

	err := v.Validate(ctx, data)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%v", ctx.Value("Limit"))
}
