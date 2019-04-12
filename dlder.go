package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type mapWithSync struct { //TODO: try sync.Map https://habr.com/ru/post/338718/
	lx  sync.Mutex
	rlx sync.RWMutex // read mutex
	mp  map[string]struct{}
}

func main() {
	start := time.Now()
	tp := flag.String("t", "", "Package to download with all requered packs")
	flag.Parse()

	// pathToDl := "/home/thevan/goDev/go-dl-debs/paks"

	resultDebMapPoiner := &mapWithSync{mp: make(map[string]struct{})}
	auxiliaryMapPointer := &mapWithSync{mp: make(map[string]struct{})}
	auxiliaryMapPointer.mp[*tp] = struct{}{} // insert target package to auxiliary map

	wg := new(sync.WaitGroup)

	for {
		if len(auxiliaryMapPointer.mp) != 0 { // LOCK??

			removePacksFromAuxiliaryMap(resultDebMapPoiner, auxiliaryMapPointer)
			newTargetPack := chooseNewPack(auxiliaryMapPointer)

			if newTargetPack != "" {
				resultDebMapPoiner.lx.Lock()
				resultDebMapPoiner.mp[newTargetPack] = struct{}{}
				resultDebMapPoiner.lx.Unlock()
				wg.Add(1)
				go lpd(newTargetPack, resultDebMapPoiner, auxiliaryMapPointer, wg) //go
			}
		} else {
			wg.Wait()
			if len(auxiliaryMapPointer.mp) == 0 { // if after alll threads still no new packs
				break
			}
		}
	}

	fmt.Println(len(resultDebMapPoiner.mp))

	// for k := range resultDebMapPoiner.mp {
	// 	fmt.Println(k) //download)
	// }

	// folderName := "packages_for_" + pack //make new folder end move to it
	// os.Mkdir(folderName, 0700)
	// ex, _ := os.Executable()
	// exPath := filepath.Dir(ex)
	// packagesFullPath := exPath + "/" + folderName
	// err := os.Chdir(packagesFullPath)

	// if err != nil {
	// 	log.Fatal("Cant enter directory: ", packagesFullPath)
	// }

	// for keyPack := range packagesMap { // download all packages
	// 	andso, _ := exec.Command("apt", "download", keyPack).Output()
	// 	fmt.Println(string(andso))
	// }
	fmt.Println(time.Since(start))
	//1m8.379200351s
	//13.683422504s
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
		panic(err) // TODO: don't panic? Retry?
	}
	resArrStr := strings.Split(string(ou), "\n")

	for i, d := range resArrStr {
		if strings.Contains(d, "PreDepends") || strings.Contains(d, "Depends") {
			strF := strings.Fields(d)
			p := strF[len(strF)-1]
			if strings.Contains(p, "<") && len(resArrStr) > 1 {
				wg2.Add(1)
				go takeSeveralPacks(resArrStr, i, secondMap, wg2)
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
