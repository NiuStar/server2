package STNet

import (
	"reflect"
	"fmt"
	"github.com/NiuStar/xsql3/Type"
	"github.com/NiuStar/json"
	//"github.com/NiuStar/xsql3"
)

//如果需要进行参数验证，必须继承该结构体
type IParameter struct {
	verValue map[string]*iParameterVer
}

func (ip *IParameter)JsonSring() string {

	body,_ := json.MarshalIndent(ip,"","\t")
	return string(body)
}

func (ip *IParameter)CountAll(t Type.DBOperation) Type.TableType {
	t1 := &Type.Int{Names:"count(*)"}
	//t1.SetTableName(t.TableName())
	return t1
}

func (ip *IParameter)Count(t Type.TableType) Type.TableType {
	t1 := &Type.Int{Names:"count(" + t.TableName() + "." + t.Name() + ")"}
	//t1.SetTableName("")
	return t1
}

func (ip *IParameter)Get() map[string]*iParameterVer {
	return ip.verValue
}
//验证参数是否为必须参数的，如果需要进行数据插入操作，对应的数据库字段名称
type iParameterVer struct {
	required bool
	paramType reflect.Type
	value reflect.Value
}

func (ip *iParameterVer)Value() reflect.Value {
	if ip != nil {
		return ip.value
	}
	return reflect.ValueOf(nil)
}

//验证必填参数，并对客户端传过来的参数进行赋值
func verificationParameter(ip *IParameter,list map[string]interface{}) *string {

	fmt.Println("paramesters:::::",ip.verValue["TestStr"])
	for name,par := range ip.verValue {
		if par.required {
			if list[name] == nil {
				return &name
			} else {

				if par.paramType.Kind() == reflect.Struct {
					fmt.Println("name:", name, list[name], par.paramType.Name(), par.paramType.String())
					if Type.IsTabelType(par.paramType) {
						if list[name] == nil {
							return &name
						}

					} else {
						errorName := verificationStruct(par.paramType, list[name].(map[string]interface{}))
						if errorName != nil {
							errorName1 := name + "_" + *errorName
							return &errorName1
						}
					}
				}


			}
		}
	}

	for name,value := range list {
		fmt.Println("ip.verValue[name]:",ip.verValue[name])
		if ip.verValue[name] != nil {
			ip.verValue[name].value = reflect.ValueOf(value)
		}

	}

	return nil
}

func verificationStruct(type_ reflect.Type,list map[string]interface{}) *string {

	for i:=0;i<type_.NumField();i++ {

		/*
		fieldTag := type_.Field(i).Tag
		if len(fieldTag.Get("json")) > 0 {
			if fieldTag.Get("required") == "yes" && list[fieldTag.Get("json")] == nil {
				errorName := fieldTag.Get("json")
				return &errorName
			} else if fieldTag.Get("required") == "yes" {

				fieldType := type_.Field(i).Type
				for ;reflect.Ptr == fieldType.Kind() || reflect.Interface == fieldType.Kind(); {
					fieldType = fieldType.Elem()
				}

				if fieldType.Kind() == reflect.Struct {
					errorName := verificationStruct(fieldType,list[fieldTag.Get("json")].(map[string]interface{}))
					if errorName != nil {
						errorName1 := fieldTag.Get("json") + "_" + *errorName
						return &errorName1
					}
				}
			}
		}*/

		fmt.Println(".......")
		errorName1 :=verificationField(type_.Field(i),list)
		if  verificationField(type_.Field(i),list) != nil {
			return errorName1
		}
	}
	return nil
}


func verificationField(field reflect.StructField,list map[string]interface{}) *string {

	fieldTag := field.Tag
	if len(fieldTag.Get("json")) > 0 {
		if fieldTag.Get("required") == "yes" && list[fieldTag.Get("json")] == nil {
			errorName := fieldTag.Get("json")
			return &errorName
		} else if fieldTag.Get("required") == "yes" {

			fieldType := field.Type
			for ;reflect.Ptr == fieldType.Kind() || reflect.Interface == fieldType.Kind(); {
				fieldType = fieldType.Elem()
			}

			if fieldType.Kind() == reflect.Struct {
				errorName := verificationStruct(fieldType,list[fieldTag.Get("json")].(map[string]interface{}))
				if errorName != nil {
					errorName1 := fieldTag.Get("json") + "_" + *errorName
					return &errorName1
				}
			}
		}
	}

	return nil
}