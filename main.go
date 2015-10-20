package main

import (
	"crypto/sha1"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"image/gif"
	"github.com/nfnt/resize"
	ini "gopkg.in/ini.v1"
//	sjson "github.com/bitly/go-simplejson"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"upload/model"
	"strconv"
	"strings"
	"time"
)

var engine *xorm.Engine
var inicfg *ini.File
var cfgStatus string
type MysqlCfg struct {
	UserName string
	Password string
	Port string
	DataName string
	Charset string
}

type cfg struct  {
	HttpPort string
	UploadName string
	Storage string
	UploadUrl string
	GetFileUrl string
}


func main() {
	defer func(){
		if err := recover();err!=nil{
			fmt.Println("yougecuowu")
		}
	}()
	go getCfg("config")
	var err error
	engine, err = xorm.NewEngine("mysql", "root:password@tcp(ip:3306)/cangku?charset=utf8")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/upload/list", list)
	http.HandleFunc("/upload/getfile/", getFile)
	http.HandleFunc("/upload/", work)
	for {
		http.ListenAndServe(":8080", nil)
	}



}

func getCfg(config string) {
	defer func(){
		if err := recover();err!=nil{
			fmt.Println(err)
		}
	}()
	var errcfg error
	for {
		data,err := ioutil.ReadFile(config)
		if err != nil{
			panic("config error")
		}
		sha1 := b(string(data))
		if sha1 == cfgStatus{
			continue
			dun := 10*time.Second
			time.Sleep(dun)
		}
		cfgStatus = sha1

		inicfg, errcfg = ini.Load(config)
		if errcfg != nil {
			panic("config error")
		}
//		fmt.Println(inicfg)
		mysql,err := inicfg.GetSection("mysql")
		if err != nil{
			panic("mysql config error")
		}
		key := mysql.Key("password")
		fmt.Println(key.String())
	}

}

func b(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func list(w http.ResponseWriter, r *http.Request) {
	fs := make([]model.Files, 0)
	query := r.URL.Query()
	page, _ := strconv.ParseInt(query.Get("page"), 10, 0)
	limit, _ := strconv.ParseInt(query.Get("rows"), 10, 0)
	start := (page - 1) * limit
	e := engine.Where("status=?", 1).Limit(int(limit), int(start)).Find(&fs)
	data := genData(fs)
	fmt.Println(fs)
	fmt.Println(e)

	ret := make(map[string]interface{})
	ret["records"] = getCount()
	ret["page"] = page
	ret["rows"] = data
	total := getCount() / limit
	if getCount() % limit != 0 {
		total = total + 1
	}
	ret["total"] = total
	t, _ := json.Marshal(ret)
	//	fmt.Println(t)
	w.Write(t)
}

func genData(fs []model.Files) []*model.ShowFile {
	show := make([]*model.ShowFile, 0)
	for _, v := range fs {
		t := new(model.ShowFile)
		t.Id = v.Id
		t.Addtime = time.Unix(v.Addtime, 0).Format("2006-01-02 15:04:05")
		t.Name = v.Name
		t.Addr = v.Addr
		t.Size = v.Size
		t.Type = v.Type
		show = append(show, t)
	}
	return show
}

func getCount() int64 {
	f := new(model.Files)
	total, _ := engine.Count(f)
	return total
}

func work(w http.ResponseWriter, r *http.Request) {
	//rw := []byte(r.URL.String())

	ufile, ft, e := r.FormFile("file")
	if e != nil {
		fmt.Println(e)
	}
	name := ft.Filename

	source, err := ioutil.ReadAll(ufile)
	if err != nil {
		fmt.Println(err)

	}
	hash := b(string(source))
	saveFile(source, hash[0:3] + "/" + name)
	f := new(model.Files)
	f.Name = name
	f.Addr = hash[0:3] + "/" + name
	f.Type = "txt"
	f.Addtime = time.Now().Unix()
	f.Hash = hash
	f.Size = len(source)
	f.Status = 1
	toCangku(f)
	im := "/upload/getfile/" + f.Addr

	w.Write([]byte(im))
	//	fmt.Println(r.FormFile("file"))
}

func saveFile(source []byte, title string) {
	strs := strings.Split(title, "/")
	l := len(strs)
	if l > 1 {
		dir := strings.Join(strs[0:l - 1], "/")
		createDir(dir)
	}
	tf, _ := os.OpenFile(title, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0777)
	tf.Write(source)
	fmt.Println(strs)
}

func toCangku(f *model.Files) {
	engine.Insert(f)
}

func createDir(dir string) bool {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return false
	} else {
		return true
	}
}

func getFile(w http.ResponseWriter, r *http.Request) {
	path := "/upload/getfile/"
	//	fmt.Println()
	reg := strings.Replace(r.RequestURI, path, "", -1)
	fmt.Println(reg)

	u := strings.Split(reg, "?")
	f, e := os.Open(u[0])
	if e != nil {
		w.WriteHeader(404)
		fmt.Println(e)
	}
	defer f.Close()
	query := r.URL.Query()
	iw := query.Get("w")
	ih := query.Get("h")
	if (iw != "" || ih != "") {
		wi, _ := strconv.ParseUint(iw, 10, 0)
		hi, _ := strconv.ParseUint(ih, 10, 0)
		img, x, _ := image.Decode(f)
		fmt.Println(x)
		m := resize.Resize(uint(wi), uint(hi), img, resize.Lanczos3)

		switch x {
		case "jpeg" :
			jpeg.Encode(w, m, nil)
		case "png" :
			png.Encode(w, m)
		case "git":
			gif.Encode(w,m,nil)
		}
	}else {
		s, _ := ioutil.ReadAll(f)

		w.Write(s)
	}
}
