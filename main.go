package main

import (
	"fmt"
	"time"
	"crypto/md5"
	"io"
	"strconv"
	"os"
	"html/template"
	"os/signal"
	"syscall"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/guotie/deferinit"
	"sync"
	"github.com/howeyc/fsnotify"
)

func main() {

	rt:=gin.Default()
	router(rt)
	deferinit.InitAll()
	deferinit.RunRoutines()
	go rt.Run(":8000")

timeTick:=time.NewTimer(20*time.Second)
	go func() {
	<-timeTick.C
		rt.LoadHTMLGlob("html/*")
	}()


	c := make(chan os.Signal, 1)
	// 信号处理
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	// 等待信号
	<-c
}

func router(r *gin.Engine) {
	r.LoadHTMLGlob("html/*")
	r.POST("/upload",upload)
	r.GET("/index",index)
}

func index(c *gin.Context) {
	c.HTML(http.StatusOK,"index.html",nil)
}


func upload(c *gin.Context) {
	fmt.Println("method:", c.Request.Method) //获取请求的方法
	if c.Request.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(c.Writer, token)
	} else {
		c.Request.ParseMultipartForm(32 << 20)
		file, handler, err := c.Request.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(c.Writer, "%v", handler.Header)
		f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)  // 此处假设当前目录下已存在test目录
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}





func init() {
	deferinit.AddRoutine(watchAutoMatedTask)
}

var (
	autoMatedTaskLock sync.RWMutex
	autoMatedTaskFile map[string]int = make(map[string]int)
)

/**
自动化任务创建文件夹监控
创建人:邵炜
创建时间:2016年9月5日14:47:10
*/
func watchAutoMatedTask(ch chan struct{}, wg *sync.WaitGroup) {
	fi, err := os.Stat("./test")
	if err != nil {
		fmt.Printf("watchAutoMatedTask: file data is error! err: %s \n", err.Error())
		return
	}
	if !fi.IsDir() {
		fmt.Printf("watchAutoMatedTask: message file name :%s is not defind! \n", "./test")
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("watchAutoMatedTask: fsnotify newWatcher is error! err: %s \n", err.Error())
		return
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev == nil {
					continue
				}
				fmt.Printf("watchAutoMatedTask: fsnotify watcher fileName: %s is change!  ev: %v \n", ev.Name, ev)
				if ev.IsDelete() {
					continue
				}
				autoMatedTaskLock.RLock()
				_, ok := autoMatedTaskFile[ev.Name]
				autoMatedTaskLock.RUnlock()
				if !ok {
					autoMatedTaskLock.Lock()
					autoMatedTaskFile[ev.Name] = 0
					autoMatedTaskLock.Unlock()
					go watchFileAutoMatedTask(ev.Name, func(path string) {
						fmt.Println("完成!")
					})
				}
			case err := <-watcher.Error:
				if err == nil {
					continue
				}
				fmt.Printf("watchAutoMatedTask: fsnotify watcher is error! err: %s \n", err.Error())
			}
		}
		done <- true
	}()
	err = watcher.WatchFlags("./test", fsnotify.FSN_MODIFY)
	if err != nil {
		fmt.Printf("watchAutoMatedTask watch error. userListLoadDir: %s  err: %s \n", "./test", err.Error())
	}

	// Hang so program doesn't exit
	<-ch

	/* ... do stuff ... */
	watcher.Close()
	wg.Done()
}

/**
自动化创建任务 需要监控的文件,判断文件是否上传完毕
创建人:邵炜
创建时间:2016年9月5日14:58:13
输入参数: 文件路劲
*/
func watchFileAutoMatedTask(filePath string, callBack func(string)) {
	defer func() {
		autoMatedTaskLock.Lock()
		delete(autoMatedTaskFile, filePath)
		autoMatedTaskLock.Unlock()
	}()
	tmrIntal := 10 * time.Second
	fileSaveTmr := time.NewTimer(tmrIntal)
	fileState, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("watchFileAutoMatedTask can't load file! path: %s err: %s \n", filePath, err.Error())
		return
	}
	var (
		size   = fileState.Size()
		number int64
	)
	<-fileSaveTmr.C
	for {
		fileState, err = os.Stat(filePath)
		if err != nil {
			fmt.Printf("watchFileAutoMatedTask can't load file! path: %s err: %s \n", filePath, err.Error())
			return
		}
		number = fileState.Size()
		if size == number {
			go callBack(filePath)
			return
		}
		size = number
		fileSaveTmr.Reset(tmrIntal)
		<-fileSaveTmr.C
	}
}