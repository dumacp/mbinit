/*
Utility to initialize microbackend

Usage of mbinit:
  -podName string
    	hostname POD (default "localhost")
  -step int
    	step of the chek
	 1, lock/verify
	 2, wait
	 3, unlock
	 (default 1)
  -url string
    	URL mongodb (default "mongodb://127.0.0.1:27017")
*/
package main


import (
	"time"
	"log"
	"flag"
	"github.com/globalsign/mgo"
)

var url string
var podName string
var buildVersion string
var step int

func init() {
	flag.StringVar(&url, "url", "mongodb://127.0.0.1:27017", "URL mongodb")
	flag.StringVar(&podName, "podName", "localhost", "hostname POD")
	flag.StringVar(&buildVersion, "buildVersion", "0.0.1", "Docker image build version")
	flag.IntVar(&step, "step", 1, "step of the chek\n\t 1, lock/verify\n\t 2, wait\n\t 3, unlock\n\t")
}

type Podflag struct {
	Id	int `bson:"_id"`
	Name	string
	BuildVersion	string
}

func main() {

	flag.Parse()

	dialInfo, err := mgo.ParseURL(url)
	dialInfo.Timeout = 5 * time.Second

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalln(err)
	}

	c := session.DB("test").C("initialize")


	pod := Podflag{}
	switch step {
	case 1:
		if err = c.FindId(1).One(&pod); err != nil {
			if err = c.Insert(&Podflag{Id: 1, Name: podName, BuildVersion: buildVersion}); err != nil {
				log.Fatalln(err)
			}
			return
		}

		if pod.Name == podName && pod.BuildVersion == buildVersion {
			if err = c.FindId(2).One(&pod); err != nil {
				return
			}
		}
		log.Fatalln("the register already exists")

	case 2:
		tick := time.Tick(3 * time.Second)
		for {
			select {
			case <-tick:
				if err = c.FindId(1).One(&pod); err != nil {
					log.Fatalf("lock not found %s", err)
				}
				if pod.Name == podName {
					return
				}
				if err = c.FindId(2).One(&pod); err == nil {
					return
				}
			}
		}
	case 3:
		if err = c.FindId(1).One(&pod); err != nil {
			log.Fatalln(err)
		}
		if pod.Name == podName && pod.BuildVersion == buildVersion {
			log.Printf("Remove lock: %v\n", pod.Name)
			if _, err = c.UpsertId(2, &Podflag{Id: 2, Name: podName, BuildVersion: buildVersion}); err != nil {
				log.Fatalln(err)
			}
		}
	default:
		log.Fatalln("Input valid step (1, 2 or 3)")
	}
}






