// diskio.go
package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	file_types  map[string]FileTypes
	cache_file  chan *CacheFile
	rx_tsv      *regexp.Regexp
	rx_file_ext *regexp.Regexp
)

////////////////////////////////////////////////////////////////////////////////

type FileTypes struct {
	MIME    string
	Expires int
	GZip    bool
}

////////////////////////////////////////////////////////////////////////////////

type CacheFile struct {
	Filename string
	Session  *Session
	Return   chan *[]byte
	Error    error
}

func (c *CacheFile) Get(session *Session, filename string) {
	c.Session = session
	c.Filename = filename
	cache_file <- c
}

type CacheDisk struct {
	Size uint64
	Data map[string]*[]byte // cached data
	Age  map[string]time.Time
}

func cacheFileManager() {
	debugLog("Starting cacheFileManager....")
	c := new(CacheDisk)
	cache_file = make(chan *CacheFile)
	c.Data = make(map[string]*[]byte)
	c.Age = make(map[string]time.Time)

	for {
		//debugLog("[diskio] 'cache_file' waiting for cache request")
		cache := <-cache_file

		if c.Data[cache.Filename] == nil ||
			cache.Session.Now.Sub(c.Age[cache.Filename]).Seconds() >= CORE_DISK_CACHE_TIMEOUT {

			c.Data[cache.Filename], cache.Error = readFile(cache.Session, windowsfyPath(cache.Filename))
			//c.Data[cache.Filename], cache.Error = readFile(cache.Session, cache.Filename)
			//debugLog("[diskio] [cacheFileManager] [$err]", cache.Error)
			if cache.Error == nil {
				c.Age[cache.Filename] = cache.Session.Now
			}
			debugLog("[diskio] creating cache")
			cache.Return <- c.Data[cache.Filename]

		} else {
			debugLog("[diskio] cache exists")
			cache.Return <- c.Data[cache.Filename]
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func init() {
	rx_tsv = regexCompile(`\s+`)
	rx_file_ext = regexCompile(`\.([a-zA-Z0-9]+$)`)
	file_types = make(map[string]FileTypes)
}

func readTemplate(session *Session, filename string, output *string) {
	if PWD_TEMPLATES_PATH != "" {
		b, err := ioutil.ReadFile(PWD_TEMPLATES_PATH + filename + ".html")
		failOnErr(err, "readTemplate")
		*output = string(b)
	} else {
		dbSelectRow(session, true, dbQueryRow(session, SQL_LAYOUT, filename), output)
	}
}

func readTemplateP(session *Session, filename string, output string) {
	readTemplate(session, filename, &output)
}

func readConf(filename string) (config [][]string) {
	b, err := ioutil.ReadFile(PWD_CONFIG_PATH + filename + ".conf")
	failOnErr(err, "readConf")

	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		config = append(config, rx_tsv.Split(l, -1))
	}

	return
}

/* Windows support is experimental and untested */
func readFile(session *Session, filename string) (b *[]byte, err error) {
	//debugLog("[diskio] [readFile]", filename)
	var out []byte
	exe := new(SysExecb)
	ext := getExt(filename)

	var fileNamePath string
	//debugLog("#######1", filename)
	if strings.HasPrefix(filename, "/uploads/") {
		fileNamePath = PWD_WRITABLE_PATH + strings.Replace(filename, "/uploads/", "/", 1)
	} else {
		fileNamePath = PWD_WEB_CONTENT_PATH + filename
	}
	//debugLog("#######2", fileNamePath)

	if (ext == "js" || ext == "css") && CORE_MINIFY_JS_CSS != "" {
		s_params := strings.Replace(CORE_MINIFY_JS_CSS, "$file", fileNamePath, -1)
		a_params := strings.Split(s_params, " ")
		cmd := a_params[0]
		params := a_params[1:]
		exe = SystemExecuteSTDOUTb(session, true, cmd, params...)
	}
	if exe.err != nil || len(exe.stdout) == 0 {
		out, err = ioutil.ReadFile(fileNamePath)
		return &out, err
	}
	return &exe.stdout, exe.err
}

func confFileTypes() {
	conf := readConf("file types")
	for _, line := range conf {
		if line[0] == "" {
			continue
		}
		if len(line) != 4 {
			debugLog("line in conf file skipped") //TODO: make this error message helpful
			continue
		}
		expires, _ := strconv.Atoi(line[2])
		var gzip bool
		if line[3] == "gzip" {
			gzip = true
		}
		file_types[line[0]] = FileTypes{MIME: line[1], Expires: expires, GZip: gzip}
	}
}

func getExt(filename string) string {
	ext := rx_file_ext.FindAllStringSubmatch(filename, 1)
	if len(ext) == 0 {
		return ""
	}
	return strings.ToLower(ext[0][1])
}

func pageWebServerMode(session *Session) {
	file := new(CacheFile)
	file.Return = make(chan *[]byte)

	go file.Get(session, session.r.URL.Path)
	b := <-file.Return
	//debugLog("channel returned")
	session.ResponseSize = len(*b)

	if file.Error != nil {
		session.Page.Section = &session.Special
		page404(session, "page")

		return
	}

	ext := getExt(session.r.URL.Path)
	ft := file_types[ext]
	session.w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", ft.Expires))
	session.w.Header().Set("Content-Type", ft.MIME)
	//if ext == "ttf" || ext == "otf" || ext == "eot" || ext == "woff" || ext == "woff2" {
	session.w.Header().Set("Access-Control-Allow-Origin", "*")
	//}

	if ft.GZip && strings.Contains(session.r.Header.Get("Accept-Encoding"), "gzip") {
		session.w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(session.w)
		_, err := gz.Write(*b)
		isErr(session, err, true, "Creating gzip writer", "pageWebServerMode")
		isErr(session, gz.Close(), true, "Closing gzip writer", "pageWebServerMode")

		return
	}

	_, err := session.w.Write(*b)
	isErr(session, err, true, "HTTP writer", "pageWebServerMode")

	return
}

/*
func purgeCacheDisk() {
	var t time.Time
	//d := time.Duration{CORE_DISK_CACHE_PURGER * 1000}
	for {
		time.Sleep(60000 * time.Millisecond)
		t = time.Now()
		for val, _ := range cache_disk.age {
			if int(t.Sub(cache_disk.age[val]).Seconds()) >= CORE_DISK_CACHE_TIMEOUT {
				delete(cache_disk.data, val)
				delete(cache_disk.age, val)
				//} else {
				//    TODO: purge older cache if > cache limit
			}
		}
	}
}
*/
