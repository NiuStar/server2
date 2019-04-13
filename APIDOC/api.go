package APIDOC


import (
	"github.com/NiuStar/utils"
	"fmt"
	"strconv"
)

type HTML struct {
	name string
	path string
	content string
	toc string //目录
	apiNumber int64
}

func NewHTML(name ,path string) *HTML {

	return &HTML{name:name,path:path}
}

func (md *HTML)WriteTitle(level int,content string) {
	md.apiNumber++
	lev := strconv.FormatInt(level,10)
	number := strconv.FormatInt(md.apiNumber,10)
	md.content += "<h" + lev + "><a name=\"heaer-n" + number + "\"></a>" + number + content + "</h" + lev + ">"
	md.toc += "<p><a href=\"#heaer-n" + number + "\">" + number + content + "</a></p>"
}

func (md *MarkDown)WriteContent(content string) {
	md.content += content
}

func (md *MarkDown)WriteImportantContent(content string) {
	md.content += "**" + content + "**"
}

func (md *MarkDown)WriteForm(contents [][]string) {

	for index,list := range contents {
		md.content += "\r\n| "
		c := ""
		for _,content := range list {
			c += content
			c += " |"
		}
		if len(c) == 0 {
			md.content += " |"
		} else {
			md.content += c
		}
		if index == 0 {
			c = ""
			for _ , _ = range list {
				c += " ---------- |"
			}
			if len(c) == 0 {
				md.content += " - |"
			} else {
				md.content += c
			}
		}
	}
}

func (md *MarkDown)Save() {

	if !utils.Mkdir(md.path) {
		fmt.Println("创建文件夹失败",md.path)
		return
	}
	data := "#" + md.name + `


` + md.content
	utils.WriteToFile(md.path + "/" + md.name + ".md",data)
	fmt.Println("写入：" + md.path + "/" + md.name + ".md")

	utils.WriteToFile(md.path + "/" + md.name + ".html",string(blackfriday.Run([]byte(data), blackfriday.WithNoExtensions())))
}


func WriteDocTitle(title string) {
	//<p class="tit">海口绕城项目</p>
}
