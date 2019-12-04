package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Image struct {
	*bytes.Buffer
	Name string
}

type Server struct {
	Name string
	Url  string
}

type Result struct {
	Server      string
	Worker      int
	Image       string
	ServerUuid  string
	Start       int64 // unixnano
	Time        int64
	ServerTime1 string
	ServerTime2 string
	ServerTime3 string
	ThreadTime1 string
	ThreadTime2 string
	ThreadTime3 string
}

func main() {
	wait := flag.Duration("wait", 0, "time to wait before starting")
	tick := flag.Duration("tick", time.Hour, "time between invocations")
	parallel := flag.Int("parallel", 50, "workers to run in parallel for phase2")
	perworker := flag.Int("perworker", 10, "images to send per worker")
	imgpath := flag.String("imgs", "./images", "path to image directory")
	resultpath := flag.String("results", "./results.csv", "path to results csv")
	serverList := flag.String("servers", "./servers", "path to file with list of servers")
	flag.Parse()

	servers := parseServers(*serverList)
	log.Printf("parsed servers:\n")
	for i, s := range servers {
		log.Printf("\t%d\t%-20s\t%s\n", i, s.Name, s.Url)
	}

	imgs := loadFiles(*imgpath)
	minImgs := (*perworker) * (*parallel + 1)
	if len(imgs) < minImgs {
		log.Fatalf("expected at least %d images, got %d", minImgs, len(imgs))
	}
	log.Printf("loaded %d imgs\n", len(imgs))

	reschan := make(chan Result, 50)
	go resultWriter(*resultpath, reschan)
	log.Printf("started resultWriter\n")

	if *wait > 0 {
		log.Printf("sleeping for %d seconds before starting\n", *wait)
		time.Sleep(*wait)
	}

	cycle(servers, imgs, reschan, parallel, perworker)
	t := time.NewTicker(*tick)
	for range t.C {
		cycle(servers, imgs, reschan, parallel, perworker)
	}
}

func cycle(servers []Server, imgs []Image, reschan chan Result, parallel, perworker *int) {
	for _, server := range servers {
		log.Printf("starting phase 1 for %s\n", server.Name)
		t0 := time.Now()
		request(0, imgs[:*perworker], server, reschan, nil, nil)
		t1 := time.Now()
		log.Printf("completed phase 1 for %s in %f seconds\n", server.Name, t1.Sub(t0).Seconds())

		log.Printf("starting phase 2 for %s\n", server.Name)
		var wgstart, wgstop sync.WaitGroup
		wgstart.Add(1)
		for i := 0; i < *parallel; i++ {
			wgstop.Add(1)
			go request(i, imgs[(i+1)*(*perworker):(i+2)*(*perworker)], server, reschan, &wgstart, &wgstop)
		}
		wgstart.Done()
		wgstop.Wait()
		log.Printf("completed phase 2 for %s in %f seconds\n", server.Name, time.Since(t1).Seconds())
	}

}

func parseServers(fp string) []Server {
	var servers []Server
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		log.Fatalf("parseServers read %s err: %v", fp, err)
	}
	bl := bytes.Split(b, []byte("\n"))
	for i, b := range bl {
		bf := bytes.Fields(b)
		if len(bf) == 0 {
			continue
		} else if len(bf) != 2 {
			log.Fatalf("parseServers parse %s line %d expected 2 fields got %d", fp, i+1, len(bf))
		}
		servers = append(servers, Server{
			Name: string(bf[0]),
			Url:  string(bf[1]),
		})
	}
	return servers
}

func resultWriter(fp string, reschan chan Result) {
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("resultWriter open %s err: %v", fp, err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	for res := range reschan {
		err = w.Write([]string{
			res.Server,
			strconv.Itoa(res.Worker),
			res.Image,
			res.ServerUuid,
			strconv.FormatInt(res.Start, 10),
			strconv.FormatInt(res.Time, 10),
			res.ServerTime1,
			res.ServerTime2,
			res.ServerTime3,
			res.ThreadTime1,
			res.ThreadTime2,
			res.ThreadTime3,
		})
		if err != nil {
			log.Printf("resultWriter write to %s err: %v\n", fp, err)
		}
	}
}

func request(worker int, imgs []Image, server Server, reschan chan Result, wgstart, wgstop *sync.WaitGroup) {
	if wgstart != nil {
		wgstart.Wait()
	}
	for _, img := range imgs {
		req, err := http.NewRequest(http.MethodPost, server.Url, img.Buffer)
		if err != nil {
			log.Printf("worker %d prepare request %s for %s err: %v\n", worker, img.Name, server.Name, err)
			continue
		}
		req.Header.Add("Content-Type", "image/jpeg")
		req.Header.Add("Accept", "image/jpeg")
		t := time.Now()
		res, err := http.DefaultClient.Do(req)
		td := time.Now().Sub(t)
		if err != nil {
			log.Printf("worker %d do %s for %s err: %v\n", worker, img.Name, server.Name, err)
			continue
		}
		ioutil.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != 200 {
			log.Printf("worker %d do %s for %s status %d %s\n", worker, img.Name, server.Name, res.StatusCode, res.Status)
			continue
		}
		st := strings.Split(res.Header.Get("time"), ",")
		tt := strings.Split(res.Header.Get("thread-time"), ",")
		if len(st) != 3 || len(tt) != 3 {
			log.Printf("worker %d do %s for %s headers expected 3 got time: %d, thread-time: %d\n", worker, img.Name, server.Name, len(st), len(tt))
			continue
		}

		reschan <- Result{
			Server:      server.Name,
			Worker:      worker,
			Image:       img.Name,
			ServerUuid:  res.Header.Get("server-uuid"),
			Start:       t.UnixNano(),
			Time:        td.Nanoseconds(),
			ServerTime1: st[0],
			ServerTime2: st[1],
			ServerTime3: st[2],
			ThreadTime1: tt[0],
			ThreadTime2: tt[1],
			ThreadTime3: tt[2],
		}
	}
	if wgstop != nil {
		wgstop.Done()
	}
}

func loadFiles(dir string) []Image {
	var bufs []Image

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("loadFiles read dir: %v", err)
	}
	for _, fi := range fis {
		b, err := ioutil.ReadFile(path.Join(dir, fi.Name()))
		if err != nil {
			log.Fatalf("loadFiles read %s err: %v", fi.Name(), err)
		}
		bufs = append(bufs, Image{bytes.NewBuffer(b), fi.Name()})
	}
	return bufs
}