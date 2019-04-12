package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

//TODO: error chan

type mapWithSync struct { //TODO: try sync.Map https://habr.com/ru/post/338718/
	lx  sync.Mutex
	rlx sync.RWMutex // read mutex
	mp  map[string]struct{}
}

func main() {
	start := time.Now()
	tp := flag.String("pc", "", "Package to download with all requered packs")
	dl := flag.String("dl", "n", "Dowload all packs, or just show. May be `y` or `n`")
	destFolder := flag.String("fl", "./packs/", "Package to download with all requered packs")
	flag.Parse()

	// pathToDl := "/home/thevan/goDev/go-dl-debs/paks"

	resultDebMapPoiner := &mapWithSync{mp: make(map[string]struct{})}
	auxiliaryMapPointer := &mapWithSync{mp: make(map[string]struct{})}
	auxiliaryMapPointer.mp[*tp] = struct{}{} // insert target package to auxiliary map

	wg := new(sync.WaitGroup)
	gotSomeErr := make(chan struct{})
	fmt.Println("Start looking 4")
	for {
		if len(auxiliaryMapPointer.mp) != 0 { // LOCK??

			removePacksFromAuxiliaryMap(resultDebMapPoiner, auxiliaryMapPointer)
			newTargetPack := chooseNewPack(auxiliaryMapPointer)

			if newTargetPack != "" {
				resultDebMapPoiner.lx.Lock()
				resultDebMapPoiner.mp[newTargetPack] = struct{}{}
				resultDebMapPoiner.lx.Unlock()
				wg.Add(1)
				go lpd(newTargetPack, resultDebMapPoiner, auxiliaryMapPointer, wg)
			}
		} else {
			wg.Wait()
			if len(auxiliaryMapPointer.mp) == 0 { // if after alll threads still no new packs
				break
			}
		}
	}

	if *dl == "y" {
		folderName := *destFolder
		os.Mkdir(folderName, 0700)
		ex, _ := os.Executable()
		exPath := filepath.Dir(ex)
		packagesFullPath := exPath + "/" + folderName
		err := os.Chdir(packagesFullPath) // go to new folder
		if err != nil {
			log.Fatal("Cant enter directory: ", packagesFullPath) //FIXME: to err chan
		}

		wg3 := new(sync.WaitGroup)
		for p := range resultDebMapPoiner.mp {
			wg3.Add(1)
			go downloadPack(p, wg3) // go
		}
		wg3.Wait()

	}

	fmt.Printf("End, total number of packages: %v\n", len(resultDebMapPoiner.mp))
	select {
	case <-gotSomeErr:
		fmt.Println("have some problems")
		// range
	}

	fmt.Println(time.Since(start))

}

func removePacksFromAuxiliaryMap(mainMap, secondMap *mapWithSync) {
	secondMap.lx.Lock()
	mainMap.rlx.Lock()
	for k := range secondMap.mp {
		if _, isExist := mainMap.mp[k]; isExist {
			delete(secondMap.mp, k)
		}
	}
	mainMap.rlx.Unlock()
	secondMap.lx.Unlock()
	return
}

func chooseNewPack(secondMap *mapWithSync) string {
	var newTargetPack string
	secondMap.lx.Lock()
	for k := range secondMap.mp {
		newTargetPack = k
		delete(secondMap.mp, k)
		break
	}
	secondMap.lx.Unlock()

	return newTargetPack
}

func lpd(packName string, mainMap, secondMap *mapWithSync, wg *sync.WaitGroup) {
	wg2 := new(sync.WaitGroup)
	ou, err := exec.Command("apt-cache", "depends", packName).Output()
	if err != nil {
		panic(err) //FIXME: to err chan
	}
	resArrStr := strings.Split(string(ou), "\n")

	for i, d := range resArrStr {
		if strings.Contains(d, "PreDepends") || strings.Contains(d, "Depends") {
			strF := strings.Fields(d)
			p := strF[len(strF)-1]
			if strings.Contains(p, "<") && len(resArrStr) > 1 {
				wg2.Add(1)
				takeSeveralPacks(resArrStr, i, secondMap, wg2) //go
			} else {
				secondMap.lx.Lock()
				secondMap.mp[p] = struct{}{}
				secondMap.lx.Unlock()
			}
		}
	}
	wg2.Wait()
	wg.Done()
	return
}

func takeSeveralPacks(resArrStr []string, i int, secondMap *mapWithSync, wg2 *sync.WaitGroup) {
	i++
	if !strings.Contains(resArrStr[i], ":") && len(resArrStr[i]) > 0 { // Sometimes apt-cache depends gives strange slice
		rawDepP := strings.Fields(resArrStr[i]) // Slice needed for easy cleaning of gaps
		dp := rawDepP[len(rawDepP)-1]
		secondMap.lx.Lock()
		secondMap.mp[dp] = struct{}{}
		secondMap.lx.Unlock()
		wg2.Add(1)
		takeSeveralPacks(resArrStr, i, secondMap, wg2)
	}
	wg2.Done()
	return
}

func downloadPack(p string, wg3 *sync.WaitGroup) {
	var exCode int
	cmd := exec.Command("apt", "download", p)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Can't execute apt download %v, got error: %v", p, err) //FIXME: to err chan
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exCode = status.ExitStatus()
			}
		}
	}

	if exCode != 0 {
		fmt.Printf("Can't download package %v", p) //FIXME: to err chan
	}

	wg3.Done()
}
