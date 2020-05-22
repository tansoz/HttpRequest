package HttpRequest

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type RequestData struct{
	Header map[string]string
	Method string
	Url *url.URL
	Cookie []*http.Cookie
	Version string
	Query map[string]string
	Timeout int64
	Host string
	conn net.Conn
	InsecureSkipVerify bool
}

func NewRequest(method string,request_url string)*RequestData{

	if tmpUrl,err := url.Parse(request_url);err == nil{

		rd := &RequestData{
			Url: tmpUrl,
			Method: strings.ToUpper(method),
			Header: make(map[string]string),
			Query: make(map[string]string),
		}
		rd.Version = "1.0"
		rd.Timeout = 15
		rd.InsecureSkipVerify = false

		// 默认头部信息
		rd.SetHeader("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
		rd.SetHeader("X-Power-By","Tansoz Studio")


		// 处理目标IP地址和端口号
		port := ":80"
		if rd.Url.Port()=="" && rd.Url.Scheme == "https" {

			port = ":443"
		}else if rd.Url.Port()!="" {

			port = ":"+rd.Url.Port()
		}

		rd.Host = rd.Url.Host+port

		for k,i := range rd.Url.Query(){

			rd.SetQuery(k,strings.Join(i,""))

		}

		return rd
	}

	return nil

}

func (this *RequestData)SetHeader(k string,v string)*RequestData{

	this.Header[k] = v

	return this
}

func (this *RequestData)SetCookies(cookies []*http.Cookie)*RequestData{

	this.Cookie = cookies

	return this
}

func (this *RequestData)SetHost(host string)*RequestData{

	this.Host = host

	return this
}

func (this *RequestData)SetInsecureSkipVerify(v bool)*RequestData{

	this.InsecureSkipVerify = v

	return this
}


func (this *RequestData)SetQuery(k string,v string)*RequestData{

	this.Query[k] = v

	return this
}

func (this *RequestData)SetTimeout(timeout int64)*RequestData{

	this.Timeout = timeout

	return this
}

func (this *RequestData)EncodeQuery()string{

	tmp := ""
	for k,i := range this.Query{
		tmp += "&"+url.QueryEscape(k)+"="+url.QueryEscape(i)
	}
	if tmp != ""{

		return "?"+tmp[1:]
	}
	return tmp
}

func (this *RequestData)Encode()string{
	header := "";
	path := "/"
	if this.Url.Path != ""{
		path = this.Url.Path
	}
	header += this.Method + " "+path+this.EncodeQuery()+" HTTP/"+this.Version+"\r\n"
	header += "Host: "+this.Url.Host+"\r\n"

	for k,i := range this.Header{

		header += k+": "+i+"\r\n"

	}

	return header
}
func (this *RequestData)Do()(*http.Response,error){
	if conn,err := this.connect();err == nil {
		if _,err = conn.Write([]byte(this.Encode()+"\r\n"));err == nil{
			req := &http.Request{Method:this.Method}
			return http.ReadResponse(bufio.NewReader(conn), req)
		}
		return nil,err
	}else{

		return nil,err
	}
}

func (this *RequestData)DoWithBody(body *BodyData)(*http.Response,error){

	if body == nil{
		return this.Do()
	}

	if conn,err := this.connect();err == nil {
		if _,err = conn.Write([]byte(this.Encode()));err == nil{
			if err = body.send(conn);err == nil{

				req := &http.Request{Method:this.Method}
				return http.ReadResponse(bufio.NewReader(conn), req)
			}
		}
		return nil,err
	}else{

		return nil,err
	}

}

func (this *RequestData)GetConnection()net.Conn{
	return this.conn
}

func (this *RequestData)Send(body *BodyData)(net.Conn,error){

	var(
		conn net.Conn
		err error
	)

	if conn,err = this.connect();err == nil{

		if body == nil {

			if _,err = conn.Write([]byte(this.Encode()+"\r\n"));err == nil{
				return conn,nil
			}
		}else{
			if _,err = conn.Write([]byte(this.Encode()));err == nil{
				if err = body.send(conn);err == nil{
					return conn,nil
				}
			}
		}
	}
	return nil,err
}

func (this *RequestData)Connect(conn net.Conn,body *BodyData)(*http.Response,error){

	if body == nil {
		if _,err := conn.Write([]byte(this.Encode()+"\r\n"));err == nil{
			req := &http.Request{Method:this.Method}
			return http.ReadResponse(bufio.NewReader(conn), req)
		}else{

			return nil,err
		}
	}else{
		if _,err := conn.Write([]byte(this.Encode()));err == nil{
			if err = body.send(conn);err == nil{

				req := &http.Request{Method:this.Method}
				return http.ReadResponse(bufio.NewReader(conn), req)
			}else{
				return nil,err
			}
		}else{
			return nil,err
		}
	}
}

func (this *RequestData)connect()(net.Conn,error){

	var(
		conn net.Conn
		err error
	)

	if this.Url.Scheme == "http"{
		conn,err = net.Dial("tcp",this.Host)
	}else{
		config := &tls.Config{
			InsecureSkipVerify:this.InsecureSkipVerify,
		}
		conn,err = tls.Dial("tcp",this.Host,config)
	}
	if err == nil{
		//conn.SetDeadline(time.Now().Add(time.Duration(this.Timeout * 1000000000)))
		return conn,nil
	}

	return nil,err


}
