// request.go
package request

import (
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

// 调试模式开关变量
var DEBUG bool

// 映射变量，字符串变量为key，interface{}类型变量为value
type Param map[string]interface{}

// request结构体，用于存储更丰富的request内容
type request struct {
	Url        string
	Params     Param
	Method     string
	Status     string
	StatusCode int
	Header     map[string][]string
	Body       string
	Response   *http.Response
	Proto      string
	Host       string
	URL        *url.URL
}

// request结构体指针变量，内部使用
var req *request

// http.Client指针变量，用于执行http请求
var client *http.Client

// auth使用的用户名和密码变量
var strUserName string = "docker"
var strPassword string = "27MTjlJyZWD0XxLf7C_SxOLlYpaprdzURn-Ec10Ew-U"

// TLS使用的认证文件的路径
var strCrtfilePath string = "/etc/docker/registry/certs/domain.crt"

// 初始化函数，在调用registry包时将优先执行
func init() {
	// 测试使用，查看是否调用registry/request包
	fmt.Printf("<=== initial registry/request package ===> \r\n")

	///////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////
}

// 创建http.Client对象
func CreateHttpClient() bool {
	bOK, pool := loadCA(strCrtfilePath)
	if !bOK {
		return false
	} else {
		// create http.Transport for Client
		transp := &http.Transport{
			TLSClientConfig:    &tls.Config{RootCAs: pool},
			DisableCompression: true,
		}
		// create http.Client
		client = &http.Client{Transport: transp}

		return true
	}
}

// loadCA is used to load a crt file， and add in pool
func loadCA(caFile string) (bool, *x509.CertPool) {
	var bOK bool

	pool := x509.NewCertPool()
	if ca, e := ioutil.ReadFile(caFile); e != nil {
		bOK = false
		if DEBUG {
			log.Println("ReadFile: ", e)
		}
	} else {
		// add to pool
		bOK = pool.AppendCertsFromPEM(ca)
	}
	return bOK, pool
}

// 内部GET方法
func (r *request) get(url string, params Param) (rest *request, err error) {
	return r.do(url, params, "GET")
}

// 内部HEAD方法
func (r *request) head(url string, params Param) (rest *request, err error) {
	return r.do(url, params, "HEAD")
}

// 内部POST方法
func (r *request) post(url string, params Param) (rest *request, err error) {
	return r.do(url, params, "POST")
}

// 内部PUT方法
func (r *request) put(url string, params Param) (rest *request, err error) {
	return r.do(url, params, "PUT")
}

// 内部DELETE方法
func (r *request) delete(url string, params Param) (rest *request, err error) {
	return r.do(url, params, "DELETE")
}

// 核心do request方法
func (r *request) do(url string, params Param, method string) (rest *request, err error) {
	// 生成NewRequest请求
	reqs, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// 添加header
	// reqs.Header.Add("Authorization", "token_value")
	// reqs.Header.Add("Content-Type", "text/plain; charset=UTF-8")
	// reqs.Header.Add("User-Agent", "Go-http-client/1.14")
	// reqs.Header.Add("Transfer-Encoding", "chunked")
	// reqs.Header.Add("Accept-Encoding", "gzip, deflate")
	reqs.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// 设置auth用户名和密码
	reqs.SetBasicAuth(strUserName, strPassword)

	// 设置参数
	// q := reqs.URL.Query()
	// for k, v := range params {
	// 	q.Add(k, fmt.Sprint(v))
	// }
	// reqs.URL.RawQuery = q.Encode()

	// 执行http.client的Do操作
	res, err := client.Do(reqs)
	if err != nil {
		return nil, err
	}
	// 延迟关闭response.Body
	defer res.Body.Close()

	// 分析response状态值，并进行部分内容的处理
	var body string
	if res.StatusCode == 200 {
		switch res.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ := gzip.NewReader(res.Body)
			for {
				buf := make([]byte, 1024)
				n, err := reader.Read(buf)
				if err != nil && err != io.EOF {
					panic(err)
				}
				if n == 0 {
					break
				}
				body += string(buf)
			}
		default:
			bodyByte, _ := ioutil.ReadAll(res.Body)
			body = string(bodyByte)
		}
	} else {
		bodyByte, _ := ioutil.ReadAll(res.Body)
		body = string(bodyByte)
	}

	// 生成返回值
	rest = &request{
		Url:        url,
		Params:     params,
		Method:     method,
		Body:       body,
		Header:     reqs.Header,
		Response:   res,
		Proto:      reqs.Proto,
		Host:       reqs.Host,
		URL:        reqs.URL,
		Status:     res.Status,
		StatusCode: res.StatusCode,
	}

	// 如果是调试模式，则显示rest详细内容
	if DEBUG || (res.StatusCode >= 400 && res.StatusCode < 500) {
		rest.Dump()
	}

	return rest, nil
}

// 以堆叠的方式显示request中的内容
func (r *request) Dump() {
	// 输出request信息
	fmt.Println("----------------------------------------------------")
	fmt.Println(r.Method, r.Proto)
	fmt.Println("Host", ":", r.Host)
	fmt.Println("URL", ":", r.URL)
	if len(r.URL.RawQuery) > 0 {
		fmt.Println("RawQuery", ":", r.URL.RawQuery)
	}
	// 循环输出request header的内容
	for key, val := range r.Header {
		fmt.Println(key, ":", val)
	}

	// 输出response信息
	fmt.Println("----------------------------------------------------")
	fmt.Println("Status", ":", r.Status)
	// 循环输出response header的内容，目前返回的Date内容中时间为格林威治时间
	for key, val := range r.Response.Header {
		fmt.Println(key, ":", val)
	}

	// 输出Body内容信息
	fmt.Println("----------------------------------------------------")
	fmt.Println(r.Body)
}

///////////////////////////////////////////////////////////////////////
// 外部可直接使用的方法

// 设定crt文件的绝对路径和名称
func SetCrtFilePath(crtfile string) (bool, error) {
	// 判断crt文件是否存在，如果存在则赋值病返回true
	_, err := os.Stat(crtfile)
	if err == nil {
		strCrtfilePath = crtfile
		return true, nil
	}

	// 如果不存在，则返回false
	if os.IsNotExist(err) {
		err = fmt.Errorf("%v is not exist", crtfile)
		return false, nil
	}

	// 其他情况返回false以及error信息
	err = fmt.Errorf("SetCrtFilePath case unknow error")
	return false, err
}

// 设定auth使用的用户名和密码
func SetAuthorizationInformation(usr string, pw string) {
	strUserName = usr
	strPassword = pw
}

// 获取auth使用的用户名和密码
func GetAuthorizationInformation() (usr, pw string) {
	usr = strUserName
	pw = strPassword

	return
}

// GET
func Get(url string, params Param) (rest *request, err error) {
	return req.get(url, params)
}

// HEAD
func Head(url string, params Param) (rest *request, err error) {
	return req.head(url, params)
}

// POST
func Post(url string, params Param) (rest *request, err error) {
	return req.post(url, params)
}

// PUT
func Put(url string, params Param) (rest *request, err error) {
	return req.put(url, params)
}

// DELETE
func Delete(url string, params Param) (rest *request, err error) {
	return req.delete(url, params)
}
