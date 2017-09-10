// *******************************************************************************************
// ** rotators.js
// **
// ** This is the main Client Application for ARSCTL. 
// ** 
// **
// ** (c) Tobias Wellnitz (DH1TW), 2015
// ********************************************************************************************


var selectedRotator; 
var rotators = {};

function compare(a,b) {
  if (a.order < b.order)
     return -1;
  if (a.order > b.order)
    return 1;
  return 0;
}

function setRotatorBtnEvntHandler(){
    //install event handler to re-draw canvas when smth has changed
    $("input[name=selected_rotator]").change(function(){
        selectedRotator = $(this).attr("id"); //update global variable
        if (rotators[selectedRotator]){ //when heading data available, then update heading
            draw_heading(rotators[selectedRotator]);                
        }
    })        
    
}

heading_options_small = {
    scale: 0.7, 
    font: "normal 9pt Arial",
    color: "#FFF",
    needleColor: "red",
    needleColorRing: "yellow",
    lineWidth: 3,
    drawDigits: false
}

// heading_options_large = {
//     scale: 2, 
//     font: "normal 14pt Arial",
//     color: "#FFF",
//     needleColor: "red",
//     needleColorRing: "yellow",
//     lineWidth: 3
// }

// indicatorOptions = {
//     scale: 2, 
//     font: "normal 14pt Arial",
//     color: "yellow",
//     lineWidth: 3
// }

// function draw_heading(heading, preSet, drawNeedle) {
//     //wrapper function to draw heading canvas
    
// 	// Grab the compass element
// 	var canvas = document.getElementById('heading');

// 	// Canvas supported?
// 	if (canvas.getContext('2d')) {
// 		var ctx = canvas.getContext('2d');
//         ctx.clearRect(0, 0, canvas.width, canvas.height);
//         draw_heading_canvas(heading, ctx, heading_options_large, drawNeedle);
        
//         if(preSet){
//             // console.log(preSet);
//             drawPositionSet(preSet, ctx, indicatorOptions, drawNeedle);
//         }
// 	}
// }




function drawMiniHeading(ctx, heading, drawNeedle) {
    //wrapper function to draw heading canvas
    
	// Grab the compass element
	var canvas = document.getElementById(ctx);

    if (canvas){
    	// Canvas supported?
    	if (canvas.getContext('2d')) {
    		var ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            draw_heading_canvas(heading, ctx, heading_options_small, drawNeedle);
    	}            
    }
}



//create leading zeros to match 3 digit format
function pad(num) {
    var s = "000" + num;
    return s.substr(s.length-3);
}



// function getMousePos(canvas, evt) {
//   var rect = canvas.getBoundingClientRect();
//   return {
//     x: evt.clientX - rect.left,
//     y: evt.clientY - rect.top
//   };
// }

function socketConnected(){
    $("#socket_lost_connection").addClass("hidden");
    $("#socket_connecting").addClass("hidden");
    $("#socket_connected").removeClass("hidden");

    setTimeout(function(){
        $("#socket_connected").addClass("hidden");
    }, 3000)
}

function socketDisconnected(){
    $("#socket_connecting").addClass("hidden");
    $("#socket_connected").addClass("hidden");
    $("#socket_lost_connection").removeClass("hidden");
    $("#main-rotator-heading").html("None");
    $("#mini-rotators").addClass("hidden");
}

function addRotator(data){
    if (data.rotator){
        //create buttons for added rotator
        var tagId = data.rotator.replace(' ', '_');
        var amountRotators = $(".mini-rotator").length;
        if ($("#"+tagId).length == 0){ //add button if it doesn't exist yet
            addMiniRotator(data);
            if(amountRotators == 0){ //first button
                selectedRotator = tagId;
                $("#"+selectedRotator+" .btn").addClass("active");
                $("#main-rotator-heading").html(data.rotator);                  
            }
            else{
                $("#mini-rotators").removeClass("hidden"); //hide MiniRoatator if only one rotator is detected
                var myRotators = $("#mini-rotators");
                var myRotatorItems = myRotators.find(".mini-rotator").sort(function(a,b){ return $(a).data('order') - $(b).data('order'); });
                myRotators.find(".mini-rotators").remove();
                myRotators.append(myRotatorItems);
            }
            // console.log("added: "+ tagId);
        }
    }
}

function addMiniRotator(data){
    var tagId = data.rotator.replace(" ", "_");
    var rotatorName = data.rotator;
    var html = "<div class='mini-rotator' data-order='" + data.order + "' id='"+ tagId + "'><button type='button' class='btn btn-primary btn-sm mini-rotator-btn' data-toggle='button' aria-pressed='false' autocomplete='off' >"+ rotatorName + "</button><canvas id='"+ tagId + "_canvas' width=70 height=70></div>"
    $("#mini-rotators").append(html);
    addRotorBtnHandler(tagId);
    drawMiniHeading(tagId+"_canvas",0,false);
}

function addRotorBtnHandler(tagId){
    btn = $("#"+tagId+" .btn");
    btn.on("click", function(){
        var btn = $("#"+tagId+" .btn");
        $(".mini-rotator-btn").removeClass("active");
        // btn.addClass("active");

        selectedRotator = tagId;
        if (rotators[selectedRotator]){ //when heading data available, then update heading
            draw_heading(rotators[selectedRotator]);
        }
        else{
            draw_heading(0,0,false);
        }
        var rotatorName = tagId.replace("_", " ");
        $("#main-rotator-heading").html(rotatorName);
    })
}    


function removeMiniRotator(data){
    if (data.rotator){
        tagId = data.rotator.replace(" ", "_");
        $("#mini-rotators #"+tagId).remove();
    }
}

function removeRotator(data){
    //remove rotator 
    if ($("#mini-rotators").children().length > 1){
        removeMiniRotator(data);
        
        tagId = data.rotator.replace(" ", "_");
        delete rotators[tagId];
        
        //set first rotator is list as new active rotator
        selectedRotator = $("#mini-rotators .mini-rotator:first-child").attr("id");
        $("#main-rotator-heading").html(selectedRotator.replace("_", " "));
        $(".mini-rotator .btn").removeClass("active");
        $("#"+selectedRotator+" .btn").addClass("active");

        //check if heading data has already been received for this rotator
        if (rotators[selectedRotator]){
            draw_heading(rotators[selectedRotator],0);                
        }
        else{
            draw_heading(0,0,false); // draw compass rose without needle
        }
        
        //hide mini-rotators when there is just one rotator left
        if ($("#mini-rotators").children().length == 1){
            $("#mini-rotators").addClass("hidden");
        }
    }
    //if just one rotator is left
    else{
        removeAllRotators();
    }
}

function removeAllRotators(){
    $("#mini-rotators").children().remove();
    rotators = {};
    selectedRotator = "undefined";
    draw_heading(0,0,false);
    $("#main-rotator-heading").html("None");
}

    
// function addCanvasClickHandler(){
// 	// Grab the compass element
// 	var canvas = document.getElementById('heading');

//     canvas.addEventListener('click', function(evt) {
//       var mousePos = getMousePos(canvas, evt);
//       var dx = mousePos.x - canvas.width / 2;
//       var dy = mousePos.y - canvas.height / 2;
//       var angle = Math.atan2(dy, dx) * (180/Math.PI) + 90;

//       if (angle < 0){
//           angle += 360;
//       }

//       angle = angle.toFixed();
//       angle = pad(angle);

//       socket.emit("command", {rotator: selectedRotator.replace("_", " "), position: angle });

//       // console.log(" angle: "+ angle);
//     }, false);
// }

function loadSocketLib(url, callback){
 
    var socket; 
    var script = document.createElement("script")
    script.type = "text/javascript";
 
    if (script.readyState) { //IE
        script.onreadystatechange = function () {
            if (script.readyState == "loaded" || script.readyState == "complete") {
                script.onreadystatechange = null;
                callback(url);
            }
        };
    } 
    else { //Others
        script.onload = function () {
            // console.log(arguments)
            callback(url);
        };
    }
    script.src = url+"/socket.io/socket.io.js";
    document.getElementsByTagName("head")[0].appendChild(script);
    return socket;
}

function openSocket(url){
    
    socket = io(url);

    socket.on("connect", function(){
        socketConnected();
        
        socket.on("reconnect", function(){
            socketConnected();
        })

        socket.on("add", function(data){
            addRotator(data);
        })

        socket.on("remove", function(data){
            removeRotator(data);
        })
        
        socket.on('heading', function(command) {
            if (command.heading.match(/[0-9]{1,3}/)){
                var heading = command.heading;
                var preSet = command.preSet;
                var rotator = command.rotator.replace(" ", "_");
                if ((heading >= 0) && (heading <=360)){
                    //copy directions of all rotators into local variable
                    rotators[rotator] = heading;
                    //update current antenna
                    if (command.rotator.replace(" ", "_") == selectedRotator){
                        draw_heading(heading, preSet);
                    }
                    drawMiniHeading(rotator+"_canvas", heading);
                }
            }        
        });
        
        socket.on("connect_error", function(error){
            console.log(error);
            removeAllRotators();
            socketDisconnected();
        })
    });
    
    socket.on("error", function(error){
        console.log(error);
    })
}

var socket; //socket connection
    
    
// $(document).ready(function(){

//     var protocol = window.location.protocol;
//     var hostname = window.location.hostname;
//     var port = $('body').data('socket_port');
//     var socketUrl = protocol + "//" + hostname + ":" + port; 
    
//     loadSocketLib(socketUrl, openSocket);
    
//     // if (($.cookie("socket-ip")) && ($.cookie("socket-port"))){
//     //     var ip = $.cookie("socket-ip");
//     //     var port = $.cookie("socket-port");
//     //     var socket_url = "http://" + ip + ":" + port;
//     //     loadSocketLib(socket_url, openSocket);
//     // }
//     // else{
//     //     $('#settingsModal').modal('show');
//     // }
//     //
//     // $("#saveModal").on("click", function(){
//     //     $.cookie("socket-ip", $("#socket-ip").val());
//     //     $.cookie("socket-port", $("#socket-port").val());
//     //     $("#settingsModal").modal("hide");
//     //     var ip = $.cookie("socket-ip");
//     //     var port = $.cookie("socket-port");
//     //     var socket_url = "http://" + ip + ":" + port;
//     //     loadSocketLib(socket_url, openSocket);
//     //    })
//     //
//     // $(".mynavbar .settings").on("click", function(){
//     //     $("#socket-ip").val($.cookie("socket-ip"));
//     //     $("#socket-port").val($.cookie("socket-port"));
//     //     $('#settingsModal').modal('show');
//     // })
    
//     draw_heading(0,0,false);
//     addCanvasClickHandler();
// })