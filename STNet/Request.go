package STNet

import (
	"github.com/NiuStar/json"
	"encoding/xml"
	"net/http"
	"io/ioutil"
	"reflect"
	"github.com/NiuStar/server2/gin"
	"strconv"
	. "github.com/NiuStar/reflect"
	"fmt"
	"strings"
	json2 "encoding/json"
	. "github.com/NiuStar/xsql3/Type"
	"github.com/NiuStar/server2/markdown"
	//"github.com/NiuStar/utils"
	"runtime"

)

var apis = markdown.APIMD{}

const IParameterName  = "IParameter"
const HandlerName  = "Handler"
const ResponseTypeName = "ResponseType"
const IResultName = "IResult"
const OP = "op"




//要求返回值结构体里必须有这个，才可以自动解析出给客户端的数据格式
type IResult struct {
	ResponseType ResponseType
}

func (result *IResult)SetResponseType(t ResponseType) {
	result.ResponseType = t
}

type Request struct {
	body string
	response interface{}
}
//对客户端来源数据进行解析，这样子只支持了Content-Type为application/json的数据
func (req *Request)DealReq( r *http.Request) error {
	s, err := ioutil.ReadAll(r.Body) //把  body 内容读入字符串 s
	req.body = string(s)
	return err
}

//根据设定的传参结构体及返回值的结构体及返回客户端的数据格式，进行自动解析
func (req *Request)ParseParam(param IHandler,parameter IParameter,handlerName string) error {

	t := reflect.ValueOf(param)
	v := reflect.New(t.Elem().Type()).Interface()

	fmt.Println("t.type:", reflect.TypeOf(&v))
	err := json.Unmarshal([]byte(req.body),v.(IHandler))
	t.Elem().Set(reflect.ValueOf(v).Elem())

	if err != nil {
		return err
	}

	t.Elem().FieldByName(IParameterName).Set(reflect.ValueOf(parameter))

	//req.response = param.Handler()
	results := t.MethodByName(handlerName).Call([]reflect.Value{})

	//return nil

	//函数返回值是否为空，空的话就代表处理客户端请求成功
	if results[0].IsNil() || !results[0].IsValid() {
		req.response = nil
	} else {
		error := reflect.New( results[0].Type()).Interface()
		error_t := reflect.ValueOf(error)
		error_t.Elem().Set(results[0])
		req.response = error
		//fmt.Println("result:",req.response)
	}
	return nil
}

type reqAndResponse struct {
	funcName string
	request *reflect.Type
	paramesters *IParameter
}

var routers = make(map[string]*reqAndResponse)

func RegisterOver() {
	apis.Write()
}

//注册路由，根据路由来设置对应的接收参数和返回值
func Register(op string,request IHandler) {

	RegisterFuncByName(op,request,"Handler")
}

//注册路由，根据路由来设置对应的接收参数和返回值
func RegisterFunc(op string,request IHandler,method func() interface{}) {

	//fmt.Println("reflect.TypeOf(request).Name():",reflect.TypeOf(request).Name())

	apis.Add(method)

	fn := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()

	fmt.Println("fn:",fn)

	name := fn[strings.LastIndex(fn,".") + 1:strings.LastIndex(fn,"-")]

	RegisterFuncByName(op,request,name)
}

//注册路由，根据路由来设置对应的接收参数和返回值
func RegisterFuncByName(op string,request IHandler,methodName string) {

	req := GetReflectType(request)

	paramesters := &IParameter{verValue:make(map[string]*iParameterVer)}
	for i:=0;i<req.NumField();i++ {
		fieldTag := req.Field(i).Tag
		if len(fieldTag.Get("json")) > 0 {
			paramesters.verValue[fieldTag.Get("json")] = &iParameterVer{required:fieldTag.Get("required")=="yes",paramType:req.Field(i).Type}

			fieldType := req.Field(i).Type
			for ;reflect.Ptr == fieldType.Kind()||reflect.Interface == fieldType.Kind(); {
				fieldType = fieldType.Elem()
			}
			paramesters.verValue[fieldTag.Get("json")].paramType = fieldType
		}
	}
	//fmt.Println("paramesters:::::",paramesters.verValue)
	routers[op] = &reqAndResponse{request:&req,paramesters:paramesters,funcName:methodName}
}


//根据url中的op字段进行路由，客户端传入参数转化为用户定义的结构体进行回调，返回值将返回至客户端，根据用户的返回值类型不同进行不同的处理
//如需返回特殊的如XML、FILE两种形式，需继承IResult，并对ResponseType进行赋值，设置返回客户端的数据格式
func RequestRouter(c *gin.Context) {
	op := c.DefaultQuery(OP,"")

	fmt.Println("op method:",op)
	var request IHandler
	var paramesters IParameter
	var handlerName string
	if routers[op] != nil {
		request = reflect.New(*routers[op].request).Interface().(IHandler)
		paramesters = *routers[op].paramesters
		handlerName = routers[op].funcName
	} else {
		c.String(http.StatusNotFound,``)
		return
	}
	fmt.Println("op method:",1)
	req := Request{}
	err := req.DealReq(c.Request)
	if err != nil {
		panic(err)
	}
	{
		//在处理参数转换之前先进行必填参数验证，如果有未传参数，直接返回
		_,have := (*routers[op].request).FieldByName(IParameterName)
		if have {
			var v interface{}
			err = json2.Unmarshal([]byte(req.body),&v)
			if err != nil {
				c.String(http.StatusOK,`{"status":`+ PARAMPORSEERROR + `,"error":"json解析失败，请检查ContentType为json"}`)
				return
			}

			fmt.Println("v:",v)
			str := verificationParameter(&paramesters,v.(map[string]interface{}))
			if str != nil {
				c.String(http.StatusOK,`{"status":`+ PARAMPORSEERROR + `,"error":"必要字段丢失` + *str + `"}`)
				return
			}
		}
	}

	fmt.Println("request type:",GetReflectType(request).Name())
	err = req.ParseParam(request,paramesters,handlerName)

	if err != nil {
		c.String(http.StatusOK,`{"status":`+ PARAMPORSEERROR + `,"error":"数据转换发生异常` + err.Error() + `"}`)
		return
	}


	respValue := GetReflectValue(req.response)

	respType := respValue.Type()

	for ;reflect.Ptr == respType.Kind()||reflect.Interface == respType.Kind(); {
		respType = respType.Elem()
	}

	if respType.Kind() == reflect.Struct {
		_,have := respType.FieldByName(ResponseTypeName)
		if !have {
			body,err := json.MarshalIndent(req.response, "", "\t")
			if err != nil {
				c.String(http.StatusOK,`{"status":`+ PARAMUNMARSHALERROR + `,"error":"服务器返回值转换异常` + err.Error() + `"}`)
				return
			}
			c.String(http.StatusOK,string(body))
			return
		} else {

			responseType := respValue.FieldByName(IResultName)
			responseType = responseType.FieldByName(ResponseTypeName)
			switch responseType.Int() {

			case STRING:{
				c.String(http.StatusOK,respValue.FieldByName("Body").String())
			}
			case XML:{
				body,err := xml.MarshalIndent(req.response, "", "\t")
				if err != nil {
					c.String(http.StatusOK,`{"status":`+ PARAMUNMARSHALERROR + `,"error":"服务器返回值转换异常` + err.Error() + `"}`)
					return
				}

				c.String(http.StatusOK,string(body))
				return
			}
			case FILE:{
				c.File(respValue.FieldByName("FilePath").String())
				return
			}
			default:
				{
					body,err := json.MarshalIndent(req.response, "", "\t")
					if err != nil {
						c.String(http.StatusOK,`{"status":`+ PARAMUNMARSHALERROR + `,"error":"服务器返回值转换异常` + err.Error() + `"}`)
						return
					}

					c.String(http.StatusOK,string(body))
					return
				}
			}
		}
		return
	} else if respType.Kind() == reflect.String {

		c.String(http.StatusOK,(*(req.response.(*interface{}))).(string))
		return
	} else if respType.Kind() == reflect.Int {

		c.String(http.StatusOK,strconv.FormatInt(int64((*(req.response.(*interface{}))).(int)),10))
		return
	} else if reflect.Map == respType.Kind() {
		body,err := json.MarshalIndent(req.response, "", "\t")
		if err != nil {
			c.String(http.StatusOK,`{"status":`+ PARAMUNMARSHALERROR + `,"error":"服务器返回值转换异常` + err.Error() + `"}`)
			return
		}

		c.String(http.StatusOK,string(body))
	} else {
		fmt.Println("123456:",req.response)
		body,err := json.MarshalIndent(req.response, "", "\t")
		if err != nil {
			c.String(http.StatusOK,`{"status":`+ PARAMUNMARSHALERROR + `,"error":"服务器返回值转换异常` + err.Error() + `"}`)
			return
		}

		c.String(http.StatusOK,string(body))
	}
}