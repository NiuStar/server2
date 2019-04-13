package markdown

import (
	"github.com/NiuStar/utils"
	"strings"
	"path/filepath"
	"go/format"
	"github.com/NiuStar/markdown"
	"strconv"
	"runtime"
	"reflect"
	"os"
	"fmt"
)
//XServer2/QrCodeLogin.(*QRCode).Login-fm
//写入API接口文档

type API struct {
	name string //method name
	path string //import path
	interfaceName string
	overwrite bool
}

type APIMD struct {
	list map[string][]*API
}

func (apis *APIMD)Add(method func() interface{}) {

	if apis.list == nil {
		apis.list = make(map[string][]*API)
	}

	a := &API{}

	fn := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()

	a.name = fn[strings.LastIndex(fn,".") + 1:strings.LastIndex(fn,"-")]

	a.interfaceName = fn[strings.LastIndex(fn,"(") + 1:strings.LastIndex(fn,")")]

	a.path = fn[:strings.Index(fn,".")][:strings.LastIndex(fn,"/")]

	a.overwrite = false

	key := fn[strings.LastIndex(fn,"/") + 1:strings.Index(fn,".")]
	apis.list[key] = append(apis.list[key],a)

}

func (apis *APIMD)Write() {

	mk := markdown.NewMarkDown("test","/Volumes/macApplication/go_work/src/XServer2")
	var apiIndex int64 = 0
	for _,values := range apis.list {

		if len(values) <= 0 {
			continue
		}
		fmt.Println("value:",*values[0])
		filepath.Walk("../" + values[0].path, func(filepath string, fi os.FileInfo, err error) error {
			if nil == fi {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			//name := fi.Name()

			fmt.Println("filepath:",filepath)
			data1 := utils.ReadFileFullPath(filepath)

			if len(data1) <= 0 || !strings.HasSuffix(filepath,".go") {
				return nil
			}
			data,err := format.Source([]byte(data1))

			if err != nil {
				panic(err)
			}

			code := string(data)
			fmt.Println("code code=",code)

			for _,api := range values {

				if api.overwrite {
					continue
				}

				//fmt.Println("" + api.interfaceName + ") " + api.name + "() interface{} {")

				//fmt.Println("*QRCode) Login() interface{} {:",strings.Index(code," *QRCode) Login() interface{} {"))

				index := strings.Index(code," " + api.interfaceName + ") " + api.name + "() interface{} {")

				fmt.Println("index 1:",index)
				if index > 0 {
					api.overwrite = true

					code1 := code[:index]
					fmt.Println("code 1:",code1)
					index = strings.LastIndex(code1,"//")
					fmt.Println("index 2:",index)
					note := code1[index+2:strings.LastIndex(code1,"\n")]
					fmt.Println("note:",note)
					apiIndex++

					mk.WriteTitle(3,strconv.FormatInt(apiIndex,10) + "." + note + " " + api.name + "\r\n")

					mk.WriteCode(`
func timer(this *XServer) {

	var count int64 = 0
	//nt := int64(time.Now().Unix())
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			{
				count++
				//fmt.Println("次数:",count)
				if len(fileList) == 0 {
					//fmt.Println("len(fileList) == 0:",count)
					fileList = getNowFiles(this, utils.GetCurrPath())
				} else {
					//fmt.Println("len(fileList) != 0",count)
					l := getNowFiles(this, utils.GetCurrPath())
					//fmt.Println(".... l := getNowFiles(\"./\") ",count)
					if len(l) != len(fileList) {
						fmt.Println("len(l) != len(fileList) : ", count)
						this.restart()
						//_ = exec.Command(fmt.Sprintf("kill -SIGUSR2 %d",os.Getpid()))

						return
					}
					for _, value := range l {
						list := value.(map[string]interface{})

						if list["modtime"].(int64) != GetTimeFromList(list["name"].(string), fileList) {
							fmt.Println("list[\"modtime\"].(int64) != GetTimeFromList(list[\"name\"].(string),fileList) :", count)
							this.restart()
							//_ = exec.Command(fmt.Sprintf("kill -SIGUSR2 %d",os.Getpid()))

							return
						}
					}
				}
			}

		}
	}
	//fmt.Println("次数:end ",count)
}`,"go")
				}
			}
			return nil
		})
	}
	mk.Save()

}
