package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type statusData struct {
	Name          string
	Online        bool
	Ram           string
	Runtime       string
	OfflineServer []string
	Players       []string
	Chat          []string
}

var serversPath = "./stordPacks"
var ruinigPath = "./run"

var systemDService = "minecraft.service"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.js")
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		scanLogs()

		var data statusData
		data.Online, data.Runtime, data.Ram = getServerInfo()
		data.Name = getName()
		data.OfflineServer = listServers()
		data.Players, data.Chat = scanLogs()

		dataJson, _ := json.Marshal(data)
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
}
func startServer() {
	if !exists(ruinigPath) {
		return
	}
	cmd := exec.Command("systemctl", "start", systemDService)
	cmd.Run()
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
}

func loadServer(pack string) {
	isLoadetServer := exists(ruinigPath)
	if isLoadetServer {
		fmt.Println("ther is alrady a loadet server")
		return
	}

	os.Rename(serversPath+"/"+pack, ruinigPath)
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

func scanLogs() (playerList []string, chatList []string) {
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

	offset := 0
	for true {
		pos := strings.Index(logs[offset+1:], "[net.minecraft.server.MinecraftServer/]:")
		if pos == -1 {
			break
		}
		pos += offset + 42

		offset = strings.Index(logs[pos:], "\n") + pos
		logText := logs[pos:offset]

		if "<" == logText[:1] {
			logText = strings.Replace(logText[1:], ">", ":", 1)

			chatList = append(chatList, logText)
		} else if "left the game" == logText[len(logText)-13:] {
			playerName := logText[:len(logText)-14]
			if slices.Contains(playerList, playerName) {
				playerList = removeFromList(playerList, playerName)
			}
		} else if "joined the game" == logText[len(logText)-15:] {
			playerName := logText[:len(logText)-16]
			if !slices.Contains(playerList, playerName) {
				playerList = append(playerList, playerName)
			}
		}
	}

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
