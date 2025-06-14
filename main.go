package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

type statusData struct {
	Name    string
	Online  bool
	Ram     string
	Runtime string

	ShutdownTime string

	OfflineServer []string
	Players       []string
	Chat          []Mgs
}

type Mgs struct {
	Time string
	Name string
	Text string
}

var serversPath = "./stordPacks"
var ruinigPath = "./run"

var systemDService = "minecraft.service"

var stordData statusData
var offset int
var autoScan bool
var autoShutdown int
var autoShutdownTaget int = 259200

func main() {
	stordData.Name = getName()
	stordData.OfflineServer = listServers()
	scanLogs()

	go scanLoppes()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.js")
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		stordData.Online, stordData.Runtime, stordData.Ram = getServerInfo()

		if !autoScan {
			stordData.OfflineServer = listServers()
			scanLogs()
		}

		dataJson, _ := json.Marshal(stordData)
		w.Write(dataJson)
	})
	http.HandleFunc("/unload", func(w http.ResponseWriter, r *http.Request) {
		unloadServer()
		w.WriteHeader(200)
	})
	http.HandleFunc("/load/{pack}", func(w http.ResponseWriter, r *http.Request) {
		loadServer(r.PathValue("pack"))
		w.WriteHeader(200)
	})
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		startServer()
		w.WriteHeader(200)
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		stopServer()
		w.WriteHeader(200)
	})

	http.ListenAndServe("0.0.0.0:8080", nil)
}

func scanLoppes() {
	if autoScan == true {
		return
	} else {
		autoScan = true
		autoShutdown = autoShutdownTaget
	}
	for autoScan {
		time.Sleep(30 * time.Second)
		stordData.OfflineServer = listServers()
		scanLogs()
		if len(stordData.Players) == 0 {
			autoShutdown -= 30
			stordData.ShutdownTime = secToHumanTime(autoShutdown)
			if autoShutdown <= 0 {
				stordData.ShutdownTime = "0"
				stopServer()
			}
		} else if autoShutdown != autoShutdownTaget {
			fmt.Println("player joind reset countdown")
			stordData.ShutdownTime = "0"
			autoShutdown = autoShutdownTaget
		}
	}
	stordData.ShutdownTime = "0"
}

func secToHumanTime(sec int) (human string) {
	var found int
	found, sec = returnRemaing(sec, 60)
	human = strconv.Itoa(found)
	if sec <= 0 {
		return
	}
	found, sec = returnRemaing(sec, 3600)
	human = strconv.Itoa(found/60) + ":" + human
	if sec <= 0 {
		return
	}
	found, sec = returnRemaing(sec, 86400)
	human = strconv.Itoa(found/3600) + ":" + human
	if sec <= 0 {
		return
	}
	human = strconv.Itoa(sec/86400) + " D-" + human
	return
}

func returnRemaing(input int, taget int) (found int, remaing int) {
	found = input % taget
	remaing = input - found
	return
}

func getName() string {
	nameByte, err := os.ReadFile(ruinigPath + "/name")
	if err != nil {
		if exists(ruinigPath) {
			fmt.Println("packet-name-not-fund")
			return "name not found"
		}
		return "nill"
	}
	return string(nameByte)
}

func stopServer() {
	cmd := exec.Command("systemctl", "stop", systemDService)
	cmd.Run()
	autoScan = false
	stordData.Players = []string{}
	stordData.Chat = []Mgs{}
}
func startServer() {
	if !exists(ruinigPath) {
		return
	}
	cmd := exec.Command("systemctl", "start", systemDService)
	cmd.Run()
	go scanLoppes()
}

func getServerInfo() (online bool, runtime string, ram string) {
	cmd := exec.Command("systemctl", "status", systemDService)

	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}

	info := out.String()

	online = strings.Contains(info, "Active: active")

	if !online {
		ram = "nil"
		runtime = "offline"
		return
	}

	ram = info[strings.Index(info, "Memory: ")+8:]
	ram = ram[:strings.Index(ram, " ")]

	runtime = info[strings.Index(info, "Active:"):]
	runtime = runtime[:strings.Index(runtime, "\n")]
	runtime = runtime[:strings.Index(runtime, " ago")]
	runtime = runtime[strings.Index(runtime, ";")+2:]
	return
}

func unloadServer() {
	online, _, _ := getServerInfo()
	if online {
		fmt.Println("stoping server for unload")
		stopServer()
	}

	os.Rename(ruinigPath, serversPath+"/"+getName())
	stordData.Name = getName()
}

func loadServer(pack string) {
	isLoadetServer := exists(ruinigPath)
	if isLoadetServer {
		fmt.Println("ther is alrady a loadet server")
		return
	}

	os.Rename(serversPath+"/"+pack, ruinigPath)
	stordData.Name = getName()
}

func listServers() (names []string) {
	data, _ := os.ReadDir(serversPath)
	for _, i := range data {
		if i.IsDir() {
			names = append(names, i.Name())
		}
	}
	return
}

func scanLogs() {
	online, _, _ := getServerInfo()
	if !online {
		return
	}
	file, err := os.ReadFile(ruinigPath + "/logs/latest.log")
	if err != nil {
		fmt.Println("log file not found")
		return
	}

	logs := string(file)
	if offset > len(logs) {
		offset = 0
	} else if offset == len(logs) {
		return
	}

	for true {
		pos := strings.Index(logs[offset+1:], "[net.minecraft.server.MinecraftServer/]:")
		if pos == -1 {
			break
		}
		pos += offset + 42

		offset = strings.Index(logs[pos:], "\n") + pos
		logText := logs[pos:offset]

		if "<" == logText[:1] {
			var newMgs Mgs
			findSplit := strings.Index(logText, ">")

			newMgs.Name = logText[1:findSplit]
			newMgs.Text = logText[findSplit+1:]

			startOffLine := strings.LastIndex(logs[:offset], "\n")
			timeStart := strings.Index(logs[startOffLine:], " ") + 1 + startOffLine
			timeEnd := strings.Index(logs[startOffLine:], ".") + startOffLine

			newMgs.Time = logs[timeStart:timeEnd]
			stordData.Chat = append(stordData.Chat, newMgs)

			if !slices.Contains(stordData.Players, newMgs.Name) {
				stordData.Players = append(stordData.Players, newMgs.Name)
			}
			if len(stordData.Chat) > 30 {
				stordData.Chat = stordData.Chat[1:]
			}
		} else if "left the game" == logText[len(logText)-13:] {
			playerName := logText[:len(logText)-14]
			if slices.Contains(stordData.Players, playerName) {
				stordData.Players = removeFromList(stordData.Players, playerName)
			}
		} else if "joined the game" == logText[len(logText)-15:] {
			playerName := logText[:len(logText)-16]
			if !slices.Contains(stordData.Players, playerName) {
				stordData.Players = append(stordData.Players, playerName)
			}
		}
	}
	offset = len(logs)
	return
}

func removeFromList(s []string, i string) []string {
	s[slices.Index(s, i)] = s[len(s)-1]
	return s[:len(s)-1]
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
