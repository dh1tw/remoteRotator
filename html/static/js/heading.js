// function draw_heading_canvas(degrees, ctx, options, drawNeedle) {
//     //draw a compass like canvas to indicate antenna heading

//     if (options.drawLegend != null){
//         var drawLegend = options.drawLegend;
//     }
//     else{
//         var drawLegend = true;
//     }

//     if (drawNeedle==null){
//         var drawNeedle = true;
//     }
    
//     if (options.drawDigits != null){
//         var drawDigits = options.drawDigits;
//     }
//     else{
//         var drawDigits = true;
//     }
    
//     var cx = 100 * options.scale; //canvas x size 
//     var cy = 100 * options.scale; //canvas y size
//     var r = 45 * options.scale; //radius
//     var font = options.font;
//     var color = options.color;
//     var needleColor = options.needleColor;
//     var needleColorRing = options.needleColorRing;
//     var lineWidth = options.lineWidth;

//     var lstx = r-(r*0.22); //large tick start x
//     var lsty = -r+(r*0.22); //large tick start y
//     var letx = r; //large tick end x
//     var lety = -r; //large tick end y    
//     var sstx = r-(r*0.11); //small tick start x
//     var ssty = -r+(r*0.11); //small tick start y
//     var setx = r; //small tick end x
//     var sety = -r; //small tick end y
    
//     //outer ring
//     ctx.beginPath();
//     ctx.strokeStyle = color;
//     ctx.lineWidth = lineWidth;
//     ctx.arc(cx/2, cy/2, r,0, 2*Math.PI);
//     ctx.stroke();
//     ctx.closePath();
    
//     ctx.save();
    
//     //draw 45° and 15° ticks
//     ctx.translate(cx/2,cy/2);
//     for (i=1;i<=360;i++) {
//         ang=Math.PI/180*i;
//         sang=Math.sin(ang);
//         cang=Math.cos(ang);
//         //If modulus of divide by 45 is zero then draw an degree marker + numeral
//         if (i % 45 == 0) {
//             ctx.lineWidth=3;
//             sx=sang * lstx;
//             sy=cang * lsty;
//             ex=sang * letx;
//             ey=cang * lety;
//             ctx.beginPath();
//             ctx.moveTo(sx,sy);
//             ctx.lineTo(ex,ey);
//             ctx.stroke();
//         }    
//         //Else draw every 10deg a small degree marker
//         else if (i % 15 == 0){
//             ctx.lineWidth=1;
//             sx=sang * sstx;
//             sy=cang * ssty;
//             ex=sang * setx;
//             ey=cang * sety;
//             ctx.beginPath();
//             ctx.moveTo(sx,sy);
//             ctx.lineTo(ex,ey);
//             ctx.stroke();
//         }
//     }

//     ctx.restore();
    
//     //North East South West Labels
//     if (drawLegend){
//         ctx.font = font;
//         ctx.fillStyle = color; 
//         ctx.fillText("W",16*options.scale, (cy/2) + 5);
//         ctx.fillText("N", cx/2-6, 25*options.scale);
//         ctx.fillText("E", cx-22*options.scale, cy/2 + 5);
//         ctx.fillText("S", cx/2-5, cy-17*options.scale);
//         ctx.fillStyle = "red";
//         ctx.font = font;
//     }
    
//     if ((drawNeedle) && (drawDigits)){
//         if ((degrees < 130) || (degrees > 240)){
//             if(degrees >= 100){
//                 ctx.fillText(degrees+"°", cx/2-15, cy-30*options.scale);
//                 }
//             else if ((degrees < 100) && (degrees >= 10)){
//                 ctx.fillText(degrees+"°", cx/2-9, cy-30*options.scale);
//             }
//             else{
//                 ctx.fillText(degrees+"°", cx/2-7, cy-30*options.scale);                
//             }
//         }
//         else{
//             ctx.fillText(degrees+"°", cx/2-15, 40*options.scale);
//         }        
//     }

//     ctx.save();
    
//     if (drawNeedle){
//         //compass needle
//         ctx.translate(cx/2,cy/2);
//         ctx.rotate(degrees * Math.PI / 180 + Math.PI);
//         ctx.beginPath();
//         ctx.moveTo(-6,0);
//         ctx.lineTo(0,r);
//         ctx.lineTo(6,0);
//         // ctx.fillStyle = "#005AAB";
//         ctx.fillStyle = needleColor;
//         ctx.closePath();
//         ctx.fill();
    
//         ctx.restore();

//         //outer ring around compass needle
//         ctx.beginPath();
//         ctx.arc(cx/2, cy/2, 6,0, 2*Math.PI);
//         // ctx.fillStyle = "#005AAB";
//         ctx.fillStyle = needleColor;
//         ctx.fill();
//         ctx.closePath();

//     }



//     //inner ring around compass needle
//     ctx.translate(0,0);
//     ctx.beginPath();
//     ctx.arc(cx/2, cy/2, 4,0, 2*Math.PI);
//     // ctx.fillStyle = "#428bca";
//     ctx.fillStyle = needleColorRing;
//     ctx.fill();
//     ctx.closePath();

    
//     ctx.restore();

// }

// function drawPositionSet(degrees, ctx, options){
//     var scale = options.scale;
    
//     var cx = 100 * scale; //canvas x size 
//     var cy = 100 * scale; //canvas y size
//     var r = 45 * scale; //radius

//     var font = options.font;
//     var color = options.color;
//     var lineWidth = options.lineWidth;
    
//     var radians = degrees*Math.PI/180;
//     var outerX = cx/2 + r * Math.cos(radians);
//     var outerY = cy/2 + r * Math.sin(radians);
    
//     ctx.save()
//     ctx.translate(cx/2,cy/2);
//     ctx.rotate(degrees * Math.PI / 180 + Math.PI);
//     ctx.beginPath();
//     ctx.strokeStyle = color;
//     ctx.lineWidth = lineWidth;
//     ctx.moveTo(0,0);
//     ctx.lineTo(0, r-2*scale);
//     ctx.moveTo(-1*scale, r-5*scale);
//     ctx.lineTo(0, r-2*scale);
//     ctx.lineTo(1*scale, r-5*scale);
//     ctx.fillStyle = color;
//     // ctx.moveTo(-6,0);
//     // ctx.lineTo(0,r);
//     // ctx.lineTo(6,0);
//     ctx.closePath();
//     ctx.stroke();

//     ctx.restore();
    
//     // ctx.beginPath();
//     // ctx.strokeStyle = color;
//     // ctx.lineWidth = lineWidth;
//     // ctx.moveTo(cx/2,cy/2);
//     // ctx.lineTo(outerX, outerY);
//     // ctx.closePath();
//     // ctx.stroke();
// }

