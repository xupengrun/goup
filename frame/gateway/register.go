package gateway

import (
	"fmt"
	"reflect"
	"time"
)

var (
	apis []*API
)

func GetApiList() []*API {
	return apis
}

// RegisterAPI 格式化的返回
func RegisterAPI(group string, key, name string, handler Handler, opts ...Option) {

	req, resp, reqType := getHandlerInOutParamInfo(handler)

	// 构建接口文档
	apis = append(apis, &API{
		LineNum:  0,
		Key:      key,
		Name:     name,
		Group:    group,
		ReqType:  reqType,
		Request:  req,
		Response: resp,
	})

	apiKey := fmt.Sprintf("%s.%s", group, key)
	//  注册到中间件中
	if _, ok := apiHandlerFuncMap[apiKey]; ok {
		panic(fmt.Errorf("%s already registered", key))
	}

	handlerInfo := &HandlerInfo{
		reqType: reqType,
		handler: reflect.ValueOf(handler).Type(),
		expire:  10 * time.Minute,
	}

	for _, opt := range opts {
		opt.apply(handlerInfo)
	}

	apiHandlerFuncMap[apiKey] = handlerInfo
}

func getHandlerInOutParamInfo(handler Handler) (in, out *DTOInfo, reqType reflect.Type) {
	req, ok := reflect.ValueOf(handler).Type().FieldByName("Request")
	if !ok {
		panic("not contains Request field")
	}

	return getDTOFieldInfo(reflect.ValueOf(handler).FieldByName("Request").Type(), false),
		getDTOFieldInfo(reflect.ValueOf(handler).FieldByName("Response").Type(), false), req.Type
}

func getDTOFieldInfo(dto reflect.Type, sub bool) *DTOInfo {
	fields := make([]*FieldInfo, 0, 10)
	types := make([]*TypeInfo, 0)
	name := dto.String()
	if sub && name == "time.Time" {
		return &DTOInfo{
			fields: fields,
			types:  types,
		}
	}

	if dto.Kind() == reflect.Interface {
		return &DTOInfo{
			fields: fields,
			types:  types,
		}
	}

	for i := 0; i < dto.NumField(); i++ {
		field := dto.Field(i)
		tag := field.Tag

		if field.Anonymous {
			info := getDTOFieldInfo(field.Type, true)
			fields = append(fields, info.fields...)
			types = append(types, info.types...)
			continue
		}

		// 处理 Foo
		if field.Type.Kind() == reflect.Struct {
			info := getDTOFieldInfo(field.Type, true)
			types = append(types, info.types...)
			continue
		}

		// 处理 *Foo
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			fmt.Printf("ptr type: %s\n", field.Type.Elem().String())
			info := getDTOFieldInfo(field.Type.Elem(), true)
			types = append(types, info.types...)
		}

		// 处理 []Foo
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			info := getDTOFieldInfo(field.Type.Elem(), true)
			types = append(types, info.types...)
		}

		// 处理 []*Foo
		if field.Type.Kind() == reflect.Slice &&
			field.Type.Elem().Kind() == reflect.Ptr &&
			field.Type.Elem().Elem().Kind() == reflect.Struct {

			info := getDTOFieldInfo(field.Type.Elem().Elem(), true)
			types = append(types, info.types...)
		}

		// 处理 map[string]Foo
		if field.Type.Kind() == reflect.Map &&
			field.Type.Elem().Kind() == reflect.Struct {
			fmt.Printf("map type: %s\n", field.Type.Elem().String())

			info := getDTOFieldInfo(field.Type.Elem(), true)
			types = append(types, info.types...)
		}

		// 处理 map[string]*Foo
		if field.Type.Kind() == reflect.Map &&
			field.Type.Elem().Kind() == reflect.Ptr &&
			field.Type.Elem().Elem().Kind() == reflect.Struct {
			info := getDTOFieldInfo(field.Type.Elem().Elem(), true)
			types = append(types, info.types...)
		}

		filedInfo := FieldInfo{
			name:     tag.Get("json"),
			desc:     tag.Get("desc"),
			typ:      field.Type.String(),
			required: true,
			note:     "todo for binding",
		}

		fields = append(fields, &filedInfo)
	}

	if sub {
		typeInfo := &TypeInfo{
			name:   name,
			fields: fields,
		}
		types = append([]*TypeInfo{typeInfo}, types...)
	}

	return &DTOInfo{
		fields: fields,
		types:  types,
	}
}

// func getTypeInfo(fieldType reflect.Type) []*TypeInfo {

// 	name := fieldType.String()
// 	types := make([]*TypeInfo, 0)

// 	if name == "time.Time" {
// 		return types
// 	}

// 	fields := make([]*FieldInfo, 0, 10)

// 	for i := 0; i < fieldType.NumField(); i++ {
// 		field := fieldType.Field(i)
// 		tag := field.Tag

// 		if field.Anonymous {
// 			info := getDTOFieldInfo(field.Type)
// 			fields = append(fields, info.fields...)
// 			types = append(types, info.types...)
// 			continue
// 		}

// 		filedInfo := FieldInfo{
// 			name:     tag.Get("json"),
// 			desc:     tag.Get("desc"),
// 			typ:      field.Type.String(),
// 			required: true,
// 			note:     "todo for binding",
// 		}

// 		fields = append(fields, &filedInfo)

// 		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
// 			types = append(types, getTypeInfo(field.Type.Elem())...)
// 		}

// 		if field.Type.Kind() == reflect.Struct {
// 			info := getTypeInfo(field.Type)
// 			types = append(types, info...)
// 		}
// 	}

// 	typeInfo := &TypeInfo{
// 		name:   name,
// 		fields: fields,
// 	}
// 	types = append([]*TypeInfo{typeInfo}, types...)

// 	return types
// }
