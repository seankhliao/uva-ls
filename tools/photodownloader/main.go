package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	// "net/url"
	"os"
	// "path/filepath"
	"time"
	// "github.com/hbagdi/go-unsplash/unsplash"
)

type myTransport struct{}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("authorization", "Client-ID "+os.Getenv("UNSPLASH"))
	return http.DefaultTransport.RoundTrip(req)
}

func main() {
	// client := &http.Client{
	// 	Transport: &myTransport{},
	// }
	//
	// u := unsplash.New(client)
	// o := &unsplash.ListOpt{Page: 1, PerPage: 100}
	// if !o.Valid() {
	// 	log.Fatalln("options not valid")
	// }
	//
	// var data [][]string
	// defer func() {
	// 	f, err := os.Create("photos-deferred.csv")
	// 	if err != nil {
	// 		log.Printf("create file: %v", err)
	// 		return
	// 	}
	// 	defer f.Close()
	//
	// 	w := csv.NewWriter(f)
	// 	w.WriteAll(data)
	// 	w.Flush()
	// 	fmt.Printf("written %d records\n", len(data))
	//
	// }()

	// for len(data) < 600 {
	// 	photos, res, err := u.Photos.All(o)
	// 	if err != nil {
	// 		if _, ok := err.(*unsplash.RateLimitError); ok {
	// 			log.Println("rate limited, sleeping for 1min")
	// 			time.Sleep(1 * time.Minute)
	// 			continue
	// 		}
	// 		log.Println("other err: ", err)
	// 		return
	// 	}
	//
	// 	for _, p := range *photos {
	// 		if *p.Width > 4000 && *p.Height > 4000 {
	// 			data = append(data, []string{strconv.Itoa(*p.Width), strconv.Itoa(*p.Height), *p.ID})
	// 		}
	// 	}
	// 	fmt.Println("current count: ", len(data))
	//
	// 	if !res.HasNextPage {
	// 		break
	// 	}
	// 	o.Page = res.NextPage
	// }
	//
	// f, err = os.Create("photos.csv")
	// if err != nil {
	// 	log.Printf("create file: %v", err)
	// 	return
	// }
	// defer f.Close()
	//
	// w := csv.NewWriter(f)
	// w.WriteAll(data)
	// w.Flush()
	// fmt.Printf("written %d records\n", len(data))

	f, err := os.Open("./photos.csv")
	if err != nil {
		log.Fatalf("open file err: %v", err)
	}
	r := csv.NewReader(f)
	data, err := r.ReadAll()
	if err != nil {
		log.Fatalf("csv readall err: %v", err)
	}
	f.Close()

	// for i := 0; i < len(data); {
	// 	dl, _, err := u.Photos.DownloadLink(data[i][2])
	// 	if err != nil {
	// 		if _, ok := err.(*unsplash.RateLimitError); ok {
	// 			log.Println("rate limited, sleeping for 1min")
	// 			time.Sleep(1 * time.Minute)
	// 			continue
	// 		}
	// 		log.Println("download link other err")
	// 		i++
	// 		continue
	// 	}
	// 	data[i] = append(data[i], dl.String())
	// 	i++
	// }
	//
	// f, err = os.Create("photos-dl.csv")
	// if err != nil {
	// 	log.Printf("create file: %v", err)
	// 	return
	// }
	// defer f.Close()
	//
	// w := csv.NewWriter(f)
	// w.WriteAll(data)
	// w.Flush()
	// fmt.Printf("written %d records\n", len(data))

	os.Mkdir("./photos", 0755)
	for i := 0; i < len(data); {
		func() {
			res, err := http.Get(fmt.Sprintf("https://unsplash.com/photos/%s/download", data[i][2]))
			if err != nil {
				log.Printf("download %d err: %v, sleeping for 1 min\n", i, err)
				time.Sleep(time.Minute)
				return
			}
			defer res.Body.Close()
			if res.StatusCode < 200 || res.StatusCode > 399 {
				log.Printf("download %d status %d, sleeping for 1 min\n", i, res.StatusCode)
				time.Sleep(time.Minute)
				ioutil.ReadAll(res.Body)
				return
			}

			i++
			fp := fmt.Sprintf("./photos/%03d-%s.jpg", i, data[i][2])
			f, err := os.Create(fp)
			if err != nil {
				log.Printf("create file %d %s err: %v\n", i, fp, err)
				return
			}
			defer f.Close()
			io.Copy(f, res.Body)

			if i%10 == 0 {
				log.Println("download progress ", i)
			}
		}()
	}
}
