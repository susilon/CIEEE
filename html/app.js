var ws;      
var isconnected = false;          
var client = loadsettings();
var username = client.username;  
var members = [];      
var destination;

// generate client key pair
var PassPhrase = Math.random().toString(16).substr(2, 9);
var Bits = 512;
var prk = cryptico.generateRSAKey(PassPhrase, Bits);
var pbk = cryptico.publicKeyString(prk);        

// function to load existing configuration
function loadsettings() {                
    if (localStorage.settings != null) {
        return JSON.parse(localStorage.settings);
    } else {
        // create configuration
        username = prompt("Please enter your name");
        var id = Math.random().toString(16).substr(2, 8);
        var clientdata = { id: id, username: username };
        localStorage.settings = JSON.stringify(clientdata);            
        return clientdata;
    }
}

const createUserPill = (objData) => {        
    var msgDiv = document.createElement("li");                 
    msgDiv.innerHTML = '<i class="fa fa-user"></i> ' + objData.Username + ' <span class="notification-badge"></span>'; 
    msgDiv.id = objData.ClientId; 
    $(msgDiv).addClass("list-group-item");        
    $(msgDiv).attr("tab","#chat-" + objData.ClientId)        

    document.getElementById("members").appendChild(msgDiv);       

    $("#chatContent").append('<div class="tab-pane fade" id="chat-' + objData.ClientId + '" role="tabpanel" aria-labelledby="' + objData.ClientId + '-tab"> Chat With ' + getUserName(objData.ClientId) + '<div class="messages row mt-3"></div></div>');
}

const sendMessage = () => {
    console.log("send to : " + destination);
    var m = getUser(destination);
    if (m) {
        var chatmsg = $("#msgBox").val();
        if (chatmsg.length > 0) {
            createBaloon(destination, "You", chatmsg);                        

            if (m.publickey != null) {            
                // public key available, encrypt chat messages
                chatmsg = cryptico.encrypt(chatmsg, m.publickey, prk);
                data = {Sid:'', Did:destination, Msg:chatmsg.cipher, Cmd:'chat'};            
                ws.send(JSON.stringify(data));
                $("#msgBox").val('');
                $("#msgBox").focus();                
            } else {
                // client doesn't have public key yet, request first
                console.log ('request public key');
                xdata = {Sid:'', Did:destination, Msg:'', Cmd:'rpk'}; 
                ws.send(JSON.stringify(xdata));
            }
        }  
    }
}

const connect = (u) => {            
    ws = new WebSocket("ws://" + document.location.host + "/ws?id=" + client.id + "&username=" + client.username);

    ws.onopen = function(message) {
        // insert self in memberlist
        members.push({ClientId:client.id, Username:client.username});
        // request list user        
        ws.send(JSON.stringify({Sid:'', Did:client.id, Msg:"", Cmd:'list'}));

        isconnected = true;
        username = u;
        updateStatus();
        $("#msgBox").val("");
        $("#msgBox").focus();
    }

    ws.onmessage = function(message) {          
        console.log('recevice data from server :');
        console.log(message.data);  
        data = JSON.parse(message.data);  
        
        // message for client
        if (data.Did == client.id) {                
        if (data.Cmd == 'rpk') {
            // receive request of public key from chat partner
            console.log ('public key sent');
            xdata = {Sid:'', Did:data.Sid, Msg:pbk, Cmd:'pks'}; 
            // send public key to requester
            ws.send(JSON.stringify(xdata));
        } else if (data.Cmd == 'pks') {
            // received public key messages from chat partner
            console.log ('public key received as following used to encrypt message before send to client id ' + data.Sid);
            var m = getUser(data.Sid);
            if (m) {
                m.publickey = data.Msg;  
            }                
            console.log(m);

            // encrypted message then send to chat partner
            console.log ('send first message :' + $("#msgBox").val());
            if ($("#msgBox").val() != "") {                             
                chatmsg = cryptico.encrypt($("#msgBox").val(), m.publickey, prk);
                xdata = {Sid:'', Did:destination, Msg:chatmsg.cipher, Cmd:'chat'};                        
                ws.send(JSON.stringify(xdata));
                $("#msgBox").val('');
                $("#msgBox").focus();
            }                    
        } else if (data.Cmd == 'list') {
            // list user from server                
            objArr = JSON.parse(data.Msg);
            for (var i=0; i < objArr.length; i++) {
                objData = objArr[i];                    
                let isExists = getUser(objData.ClientId); 
                if (!isExists) {
                    members.push(objData);
                    createUserPill(objData);
                }
            }
        } else if (data.Cmd == 'chat') {                    
            console.log('encrypted message :');
            console.log(data.Msg);

            var messages = data.Msg;
            if (data.Sid != client.id) {
                // decrypt message
                var uncrypted = cryptico.decrypt(data.Msg, prk);                   
                console.log('decrypted message :');
                console.log(uncrypted.plaintext);                     
                // put on chat screen
                createBaloon(data.Sid, getUserName(data.Sid), uncrypted.plaintext);           
                
                if (destination != data.Sid) {
                    var badge = $("#"+data.Sid).find('.notification-badge');
                    badge.text('new');
                    badge.addClass("badge");
                    badge.addClass("badge-success");
                }                
            }    
        } else if (data.Cmd == 'join') {
            console.log("new user joined");                
            objData = JSON.parse(data.Msg);
            let isExists = getUser(objData.ClientId); 
            if (!isExists) {
                members.push(objData);
                createUserPill(objData);                       
            } else {
                console.log('member already exists');
            }
        } else if (data.Cmd == 'leave') {
            console.log("user leave");                
            objData = JSON.parse(data.Msg);
            let member = getUser(objData.ClientId);                 
            if (member) {
                members.splice (members.indexOf(member), 1)                    
                $("#"+objData.ClientId).remove();                    
            } else {
                console.log('cannot remove members');
            }
        }
        }           
    };

    ws.onclose = function() {
        isconnected = false;
        console.log("ws disconnected");
        document.getElementById("members").innerHTML = "";
        updateStatus();
        members = []; // reset members
        $("#members").html(""); // clear members
    }
}

const updateStatus = () => {
    document.getElementById("msgConnect").innerHTML = isconnected?"Connected as " + username + " | <a href='javascript:' id='logout'>Logout</a>":"Server Not Connected";
    if (isconnected) {
        $("#btnConnect").hide();
        $("#btnSend").removeClass('disabled');

        var msgDiv = document.createElement("li"); 
        msgDiv.innerHTML = '<i class="fa fa-user"></i> You : ' + username; 
        msgDiv.id = "self";
        $(msgDiv).addClass("list-group-item");
        $(msgDiv).addClass("bg-success");
        $(msgDiv).addClass("text-light");
        $(msgDiv).attr("tab","#home")

        document.getElementById("members").appendChild(msgDiv);
    } else {
        $("#btnConnect").text("Connect" + (username != ""?" as " + username:""));
        $("#btnConnect").show();
        $("#btnSend").addClass('disabled');
    }
}    

const askUserName = () => {
    Swal.fire({
        title: 'Your name',
        input: 'text',
        inputAttributes: {
            autocapitalize: 'off'
        },
        showCancelButton: false,
        confirmButtonText: 'Submit',
        showLoaderOnConfirm: true,
        preConfirm: (username) => {
            if (username.length >= 3 && username.length <= 7) {
                // Handle return value 
            } else {
                Swal.showValidationMessage('username must between 3 and 7 char')   
            }
        },
        allowOutsideClick: false,
    }).then((result) => {
        console.log("Connecting as " + result.value + "...");
        connect(result.value);
    });
}

const createBaloon = (id, user, message) => {
    var baloon;
    if (user == "You") {
        baloon = '<div class="col-12"><div class="badge badge-success p-2 my-balloon m-2">\
        <div class="friends-name">You</div><div class="mt-1 message">' + message + '</div></div></div>';
        } else if (user == "server") {
        baloon = '<div class="col-12" style="text-align:center;"><div class="badge badge-warning p-1 server-balloon m-2">\
        <div class="message">' + message + '</div></div></div>';
        } else {
        baloon = '<div class="col-12"><div class="badge badge-secondary p-2 friends-balloon m-2">\
        <div class="friends-name">' + user + '</div><div class="mt-1 message">' + message + '</div></div></div>';
        }

    $("#chat-" + id).find('.messages').append(baloon);
}

const getUserName = (id) => {
    user = members.find(c => c.ClientId == id);
    return user.Username; 
}

const getUser = (id) => {
    return members.find(c => c.ClientId == id);
}