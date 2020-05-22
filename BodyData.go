package HttpRequest

import (
	"encoding/base64"
	"mime"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)


const(
	BODY_URLENCODED = iota
	BODY_MULTIPART
	BODY_JSON
	BODY_XML
)

var contentType = []string{
	BODY_URLENCODED: "application/x-www-form-urlencoded;charset=utf-8",
	BODY_MULTIPART: "multipart/form-data; boundary=----WebKitFormBoundary",
	BODY_JSON: "application/json;charset=utf-8",
	BODY_XML: "text/xml",
}

type BodyData struct {
	Type int
	File map[string]*os.File
	Query map[string]string
	JSONText string
	XMLText string
}

func NewBodyData(BodyType int)*BodyData{

	bd := &BodyData{
		Type: BodyType,
		File: make(map[string]*os.File),
		Query: make(map[string]string),
	}

	return bd
}

func (this *BodyData)AddQuery(k string,v string)*BodyData{

	this.Query[k] = v

	return this
}

func (this *BodyData)AddFile(k string,filepath string)*BodyData{

	if f,err := os.Open(filepath);err == nil{
		this.File[k] = f
	}

	return this
}

func (this *BodyData)SetJSONText(json string)*BodyData{

	this.JSONText = json

	return this
}

func (this *BodyData)SetXMLText(xml string)*BodyData{

	this.XMLText = xml

	return this
}

func (this *BodyData)SetType(t int)*BodyData{

	this.Type = t

	return this
}

func base64string(v string)string{

	return base64.StdEncoding.EncodeToString([]byte(v))

}

func stringEscape(v string)string{

	index := regexp.MustCompile("\\\"|\\\\").FindAllStringIndex(v,-1)
	out := ""
	start := 0
	for _,i := range index{

		out += v[start:i[0]]
		char := v[i[0]:i[1]]
		start = i[0] + 1
		if char == "\\"{
			out += "\\\\"
		}else if char == "\""{
			out += "\\\""
		}
	}
	out += v[start:]

	return out

}

func getFileNameExtension(v string)string{
	s := strings.Split(v,".")
	return "."+s[len(s) - 1]
}

func (this *BodyData)send(conn net.Conn) error {

	content := "Content-Type: " + contentType[this.Type]
	if this.Type == BODY_MULTIPART {
		key := base64string(time.Now().String())[0:16]
		join := "------WebKitFormBoundary" + key
		content += key + "\r\n"

		output := join
		for k,i := range this.Query{
			output += "\r\nContent-Disposition: form-data; name=\""+stringEscape(k)+"\"\r\n\r\n"+i+"\r\n"+join
		}

		contentLenght := int64(len(output))

		tmpFileSet := make(map[*os.File]string)
		for k,f := range this.File{
			fileinfo,_ := f.Stat()
			mimetype := mime.TypeByExtension(getFileNameExtension(fileinfo.Name()))
			if mimetype == ""{
				mimetype = "application/octet-stream"
			}
			tmp := "\r\nContent-Disposition: form-data; name=\""+stringEscape(k)+"\"; filename=\""+stringEscape(fileinfo.Name())+"\"\r\nContent-Type: "+mimetype+"\r\n\r\n"
			contentLenght += int64(len(tmp)) + 44 + fileinfo.Size()
			tmpFileSet[f] = tmp
		}
		content += "Content-Length: "+strconv.FormatInt(contentLenght,10)+"\r\n\r\n"
		if _,write_err := conn.Write([]byte(content+output));write_err != nil {
			return write_err
		}

		for f,s := range tmpFileSet{

			if _,write_err := conn.Write([]byte(s));write_err != nil {
				return write_err
			}

			fileinfo,_ := f.Stat()
			var i int64 = 0
			size := fileinfo.Size()
			b := make([]byte,1024)
			for{
				if i >= size {
					break
				}

				num,err := f.Read(b)
				_,write_err := conn.Write(b[0:num])
				if err != nil && err.Error() != "EOF" {
					return err
				}else if write_err != nil {
					return write_err
				}

				i += int64(num)
			}
			if _,write_err := conn.Write([]byte("\r\n"+join+"--"));write_err != nil {
				return write_err
			}

		}

	}else if this.Type == BODY_URLENCODED {
		output := ""
		for k,i := range this.Query {

			output += "&"+url.QueryEscape(k)+"="+url.QueryEscape(i)
		}
		contentLength := len(output[1:])
		content += "\r\nContent-Length: "+strconv.FormatInt(int64(contentLength),10)+"\r\n\r\n"
		if _,write_err := conn.Write([]byte(content+output[1:]));write_err != nil {
			return write_err
		}
	}else if this.Type == BODY_JSON {

		content += "Content-Length: "+strconv.FormatInt(int64(len(this.JSONText)),10)+"\r\n\r\n"
		if _,write_err := conn.Write([]byte(content+this.JSONText));write_err != nil {
			return write_err
		}

	}else if this.Type == BODY_XML {

		content += "Content-Length: "+strconv.FormatInt(int64(len(this.XMLText)),10)+"\r\n\r\n"
		if _,write_err := conn.Write([]byte(content+this.XMLText));write_err != nil {
			return write_err
		}

	}

	return nil
}

