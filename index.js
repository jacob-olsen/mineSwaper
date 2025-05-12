var online
var offlineCount

function Start(){
    updateStatus()
    setInterval(updateStatus,10000)
}

function updateStatus(){
    fetch("/status")
    .then(res => res.json())
    .then(out =>
        updateUi(out))
    .catch(err => console.log(err));
}

function updateUi(data){
    online = data["Online"]
    if (online == true){
        document.getElementById("start").innerText = "stop"
    }else{
        document.getElementById("start").innerText = "start"
    }

    document.getElementById("name").innerText = data["Name"]
    document.getElementById("online").innerText = "runing: " + data["Online"]
    document.getElementById("runtime").innerText = "runtime: " + data["Runtime"]
    document.getElementById("ram").innerText = "RAM: " + data["Ram"]

    updateOffline(data["OfflineServer"])
    if (data["ShutdownTime"] == 0){
        updatePlayerList(data["Players"])
    }else{
        updateShutdownTimmer(data["ShutdownTime"])
    }
    
    updateChat(data["Chat"])
}

function switchStad(){
    if (online == true){
        fetch("/stop")
    }else{
        fetch("/start")
    }
    updateStatus()
}

function updateOffline(data){
    if(offlineCount != data.length){
        offlineCount = data.length
        const taget = document.getElementById("serverList")
        taget.innerHTML = ""

        data.forEach( (element) => {
            taget.innerHTML += '<div class="col-4 card"><div class="row"><p>'+element+'</p></div><div class="row"><button onclick="loadServer(\''+element+'\')">load</button></div></div>'
        });
    }
}

function updateShutdownTimmer(data){
    const taget = document.getElementById("playerList")
    taget.innerHTML = '<div class="row"><h2>server shutdown in</h2></div><div class="row"><h1>'+data+' sec</h1></div>'

}

function updatePlayerList(data){

        const taget = document.getElementById("playerList")
        taget.innerHTML = ""

        data.forEach( (element) => {
            taget.innerHTML += '<div class="col-6"><p>'+element+'</p></div>'
        });
}

function updateChat(data){

    const taget = document.getElementById("chat")
    taget.innerHTML = ""

    data.forEach( (element) => {
        taget.innerHTML += '<div class="row"><p>'+element["Time"]+" "+element["Name"]+'</p></div> <div class="row"><p>'+element["Text"]+'</p></div> <div class="row"><hr></div>'
    });

}

function loadServer(name){
    fetch("/load/"+name)
    updateStatus()
}

function unloadServer(){
    fetch("/unload")
    updateStatus()
}