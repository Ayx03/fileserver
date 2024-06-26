package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

type shareDir struct {
	dir string
}
type dirItem struct {
	N    int
	Name string
	Typ  string
}

func (s *shareDir) GetRealPath(path1 string) string {
	return path.Join(s.dir, path1)
}

var fileServ http.Handler

//GetType 0-not exist,1-file,2-dir
func (s *shareDir) GetType(realPath string) int {
	if strings.HasPrefix(realPath, "../") || strings.Contains(realPath, "/..") {
		return 0
	}
	stat1, err := os.Stat(realPath)
	if err != nil {
		log.Println(err)
		return 0
	}
	if stat1.IsDir() {
		return 2
	}
	return 1
}

func sortDirs(infos []os.FileInfo) []os.FileInfo {
	sort.Slice(infos, func(i, j int) bool {
		info1 := infos[i].(os.FileInfo)
		info2 := infos[j].(os.FileInfo)
		return strings.Compare(info1.Name(), info2.Name()) == -1
	})
	return infos
}

//ListDir 必须先确定为目录
func (s *shareDir) ListDir(realPath string) []dirItem {
	dir1, _ := os.Open(realPath)
	infos, _ := dir1.Readdir(0)
	infos = sortDirs(infos)
	res := []dirItem{}
	var typ string
	for i, v := range infos {
		tail := ""
		if v.IsDir() {
			typ = "D"
			tail = "/"
		} else {
			typ = "F"
		}
		res = append(res, dirItem{i + 1, v.Name() + tail, typ})
	}
	return res
}

var share *shareDir

const shareBase = "/share/"
const lenShareBase = len(shareBase)

func setShareDir(dir string) {
	http.HandleFunc(shareBase, handleShare)
	share = &shareDir{dir}
	//fileServ = http.StripPrefix(shareBase, http.FileServer(http.Dir(dir)))
}

func handleShare(w http.ResponseWriter, r *http.Request) {
	//fileServ.ServeHTTP(w, r)
	realPath := share.GetRealPath(r.URL.Path[lenShareBase:])
	switch share.GetType(realPath) {
	case 0:
		w.WriteHeader(404)
	case 1:
		http.ServeFile(w, r, realPath)
	case 2:
		// buf := bytes.NewBufferString(r.URL.Path + "\n")
		list1 := share.ListDir(realPath)
		// for _, v := range list1 {
		// 	buf.WriteString(fmt.Sprintf("%d . %s  %v\n", v.N, v.Typ, v.Name))
		// }
		t := template.New("")
		t.Parse(tmplDir)
		data := make(map[string]interface{})
		data["title"] = r.URL.Path
		data["links"] = list1
		w.WriteHeader(200)
		t.Execute(w, data)
	}
}
