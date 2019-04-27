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
var runLockVer string
var step int

func init() {
	flag.StringVar(&url, "url", "mongodb://127.0.0.1:27017", "URL mongodb")
	flag.StringVar(&podName, "podName", "localhost", "hostname POD")
	flag.StringVar(&buildVersion, "buildVersion", "0.0.1", "Docker image build version")
	flag.StringVar(&runLockVer, "runLockVer", "0.0.1", "run lock version")
	flag.IntVar(&step, "step", 1, "step of the chek\n\t 1, lock/verify\n\t 2, wait\n\t 3, unlock\n\t 4, indexes\n\t")
}

type Podflag struct {
	Id	int `bson:"_id"`
	Name	string
	Type	string
	BuildVersion	string
	RunLockVer	string
	TimeStamp	int64
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


	ts := time.Now().Unix()
	pod := &Podflag{}
	podLock := &Podflag{Id: 1, Name: podName, BuildVersion: buildVersion, RunLockVer: runLockVer, Type: "lock", TimeStamp: ts}
	podUnLock := &Podflag{Id: 2, Name: podName, BuildVersion: buildVersion, RunLockVer: runLockVer, Type: "unlock", TimeStamp: ts}
	podIndexes := &Podflag{Id: 3, Name: podName, BuildVersion: buildVersion, RunLockVer: runLockVer, Type: "indexes", TimeStamp: ts}

	switch step {
	case 1:
		if err = c.FindId(1).One(&pod); err != nil {
			if err = c.Insert(podLock); err != nil {
				log.Fatalln(err)
			}
			return
		}
		if pod.RunLockVer != runLockVer {
			if err = c.RemoveId(1); err == nil {
				if err = c.Insert(podLock); err != nil {
					log.Fatalln(err)
				}
				return
			}
		}
		if pod.Name == podName {
			if err = c.FindId(2).One(&pod); err != nil {
				if err == mgo.ErrNotFound {
					return
				}
				log.Fatalln(err)
			}
		}
		log.Fatalln("the register already exists")

	case 2:
		tick := time.NewTicker(3 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				if err = c.FindId(1).One(&pod); err != nil {
					log.Fatalf("lock not found %s", err)
				}
				if pod.Name == podName {
					return
				}
				if err = c.FindId(2).One(&pod); err == nil {
					if pod.RunLockVer == runLockVer {
						return
					}
				}
			}
		}
	case 3:
		if err = c.FindId(1).One(&pod); err != nil {
			log.Fatalln(err)
		}
		if pod.Name == podName {
			log.Printf("Remove lock: %v\n", pod.Name)
			if _, err = c.UpsertId(2, podUnLock); err != nil {
				log.Fatalln(err)
			}
		}
	case 4:
		if err = c.FindId(1).One(&pod); err != nil {
                        log.Fatalln(err)
                }
		if err = c.FindId(2).One(&pod); err != nil {
                        log.Fatalln(err)
		}
		if err = c.FindId(3).One(&pod); err != nil {
                        log.Println("Don't flag indexes")
			pod.TimeStamp = 0
                } else {
			if pod.RunLockVer != runLockVer {
				pod.TimeStamp = 0
			}
		}



		if ts > (pod.TimeStamp + 60*60*1) {
			if _, err = c.UpsertId(3, podIndexes); err != nil {
				log.Fatalln(err)
			}
			log.Println("Update flag indexes")
			return
		}
		log.Fatalln("don't update indexes")
	default:
		log.Fatalln("Input valid step (1, 2, 3 or 4)")
	}
}






