// $.material.init();

var vm = new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        rotators: {},
        selectedRotator: "n/a",
        hideConnectionMsg: false,
        total: 0,
    },
    beforeCreate: function(){
    },
    created: function () {
    },
    mounted: function () {
        this.canvas = this.$refs.azimuthHeading;
        this.ctx = this.canvas.getContext("2d");
        this.drawHeading(0,0,false);
        this.getRotators();
    },
    methods: {
        incrementTotal: function (){
            this.total += 1
        },
        openWebsocket: function () {
            this.ws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');
            this.ws.addEventListener('message', function (e) {
                var rotatorsMsg = JSON.parse(e.data);
                for (i = 0; i < rotatorsMsg.length; i++) {
                    newRotator = rotatorsMsg[i]
                    if (newRotator.name in this.rotators){
                        // copy values
                        rotator = this.rotators[newRotator.name]
                        if (rotator.has_azimuth){
                            rotator.azimuth = newRotator.azimuth
                            rotator.az_preset = newRotator.az_preset    
                        }
                        if (rotators.has_elevation){
                            rotator.elevation = newRotator.elevation
                            rotator.el_preset = newRotator.el_preset    
                        }

                        // update canvas
                        if (rotator.name == this.selectedRotator){
                            this.drawHeading(rotator.azimuth, rotator.az_preset, true)
                        }
                    }
                }
            }.bind(this));

            this.ws.addEventListener('open', function () {
                this.connected = true
                setTimeout(function () {
                    this.hideConnectionMsg = true;
                }.bind(this), 1500)
            }.bind(this));

            this.ws.addEventListener('close', function () {
                this.connected = false
                this.hideConnectionMsg = false;
            }.bind(this));
        },
        getRotators: function () {
            this.$http.get("/info").then(rotators => {
                rotatorsInfoMsg = JSON.parse(rotators.bodyText);
                for (i = 0; i < rotatorsInfoMsg.length; i++) {
                    // only if a rotator is not registered, add it
                    if (!(rotatorsInfoMsg[i].name in this.rotators)){
                        this.addRotator(rotatorsInfoMsg[i]);
                    }
                }
                // TBD check if a rotator has disappeared

                if (this.ws == null){
                    this.openWebsocket();
                }
            });
        },
        addRotator: function (rotator) {
            name = rotator.name;
            if (!(name in this.rotators)) {
                this.rotators[name] = rotator;
                // if this is the first rotator, select it
                if (this.selectedRotator == "n/a") {
                    this.selectedRotator = name;
                    this.addCanvasClickHandler();
                }
            }
        },
        drawHeading: function (heading, preset, drawNeedle) {
            //wrapper function to draw heading canvas

            // each time we draw something on the canvas we have to clear it
            // before drawing
            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
            this.drawHeadingCanvas(heading, this.ctx, heading_options_large, drawNeedle);

            // only draw preset arrow when the heading is >=3deg away from preset
            if (Math.abs(heading-preset) >= 3) {
                this.drawPositionSet(preset, this.ctx, indicatorOptions, drawNeedle);
            }
        },

        // add a listener for clicks on the canvas. The listener will 
        // retrieve the selected heading and send a request to the
        // server over the websocket.
        addCanvasClickHandler: function () {

            this.canvas.addEventListener('click', function (evt) {
                var mousePos = this.getMousePosition(this.canvas, evt);
                var dx = mousePos.x - this.canvas.width / 2;
                var dy = mousePos.y - this.canvas.height / 2;
                var angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90;

                if (angle < 0) {
                    angle += 360;
                }

                var msg = {
                    "name" : this.selectedRotator,
                    "has_azimuth": true,
                    "azimuth": Math.round(angle, 0),
                }
                data = JSON.stringify(msg);

                this.ws.send(data);
            }.bind(this), false);
        },
        getMousePosition: function (canvas, evt) {

            var rect = canvas.getBoundingClientRect();
            return {
                x: evt.clientX - rect.left,
                y: evt.clientY - rect.top
            };
        },
        // draw a compass like canvas to indicate the rotator heading
        // only when a rotator 
        drawHeadingCanvas(degrees, ctx, options, drawNeedle) {

            if (options.drawLegend != null) {
                var drawLegend = options.drawLegend;
            } else {
                var drawLegend = true;
            }

            if (drawNeedle == null) {
                var drawNeedle = true;
            }

            if (options.drawDigits != null) {
                var drawDigits = options.drawDigits;
            } else {
                var drawDigits = true;
            }

            var cx = 100 * options.scale; //canvas x size 
            var cy = 100 * options.scale; //canvas y size
            var r = 45 * options.scale; //radius
            var font = options.font;
            var color = options.color;
            var needleColor = options.needleColor;
            var needleColorRing = options.needleColorRing;
            var lineWidth = options.lineWidth;

            var lstx = r - (r * 0.22); //large tick start x
            var lsty = -r + (r * 0.22); //large tick start y
            var letx = r; //large tick end x
            var lety = -r; //large tick end y    
            var sstx = r - (r * 0.11); //small tick start x
            var ssty = -r + (r * 0.11); //small tick start y
            var setx = r; //small tick end x
            var sety = -r; //small tick end y

            //outer ring
            ctx.beginPath();
            ctx.strokeStyle = color;
            ctx.lineWidth = lineWidth;
            ctx.arc(cx / 2, cy / 2, r, 0, 2 * Math.PI);
            ctx.stroke();
            ctx.closePath();

            ctx.save();

            //draw 45° and 15° ticks
            ctx.translate(cx / 2, cy / 2);
            for (i = 1; i <= 360; i++) {
                ang = Math.PI / 180 * i;
                sang = Math.sin(ang);
                cang = Math.cos(ang);
                //If modulus of divide by 45 is zero then draw an degree marker + numeral
                if (i % 45 == 0) {
                    ctx.lineWidth = 3;
                    sx = sang * lstx;
                    sy = cang * lsty;
                    ex = sang * letx;
                    ey = cang * lety;
                    ctx.beginPath();
                    ctx.moveTo(sx, sy);
                    ctx.lineTo(ex, ey);
                    ctx.stroke();
                }
                //Else draw every 10deg a small degree marker
                else if (i % 15 == 0) {
                    ctx.lineWidth = 1;
                    sx = sang * sstx;
                    sy = cang * ssty;
                    ex = sang * setx;
                    ey = cang * sety;
                    ctx.beginPath();
                    ctx.moveTo(sx, sy);
                    ctx.lineTo(ex, ey);
                    ctx.stroke();
                }
            }

            ctx.restore();

            //North East South West Labels
            if (drawLegend) {
                ctx.font = font;
                ctx.fillStyle = color;
                ctx.fillText("W", 16 * options.scale, (cy / 2) + 5);
                ctx.fillText("N", cx / 2 - 6, 25 * options.scale);
                ctx.fillText("E", cx - 22 * options.scale, cy / 2 + 5);
                ctx.fillText("S", cx / 2 - 5, cy - 17 * options.scale);
                ctx.fillStyle = "red";
                ctx.font = font;
            }

            if ((drawNeedle) && (drawDigits)) {
                if ((degrees < 130) || (degrees > 240)) {
                    if (degrees >= 100) {
                        ctx.fillText(degrees + "°", cx / 2 - 15, cy - 30 * options.scale);
                    } else if ((degrees < 100) && (degrees >= 10)) {
                        ctx.fillText(degrees + "°", cx / 2 - 9, cy - 30 * options.scale);
                    } else {
                        ctx.fillText(degrees + "°", cx / 2 - 7, cy - 30 * options.scale);
                    }
                } else {
                    ctx.fillText(degrees + "°", cx / 2 - 15, 40 * options.scale);
                }
            }

            ctx.save();

            if (drawNeedle) {
                //compass needle
                ctx.translate(cx / 2, cy / 2);
                ctx.rotate(degrees * Math.PI / 180 + Math.PI);
                ctx.beginPath();
                ctx.moveTo(-6, 0);
                ctx.lineTo(0, r);
                ctx.lineTo(6, 0);
                // ctx.fillStyle = "#005AAB";
                ctx.fillStyle = needleColor;
                ctx.closePath();
                ctx.fill();

                ctx.restore();

                //outer ring around compass needle
                ctx.beginPath();
                ctx.arc(cx / 2, cy / 2, 6, 0, 2 * Math.PI);
                // ctx.fillStyle = "#005AAB";
                ctx.fillStyle = needleColor;
                ctx.fill();
                ctx.closePath();

            }

            //inner ring around compass needle
            ctx.translate(0, 0);
            ctx.beginPath();
            ctx.arc(cx / 2, cy / 2, 4, 0, 2 * Math.PI);
            // ctx.fillStyle = "#428bca";
            ctx.fillStyle = needleColorRing;
            ctx.fill();
            ctx.closePath();


            ctx.restore();

        },

        drawPositionSet: function (degrees, ctx, options) {
            var scale = options.scale;

            var cx = 100 * scale; //canvas x size 
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            var font = options.font;
            var color = options.color;
            var lineWidth = options.lineWidth;

            var radians = degrees * Math.PI / 180;
            var outerX = cx / 2 + r * Math.cos(radians);
            var outerY = cy / 2 + r * Math.sin(radians);

            ctx.save()
            ctx.translate(cx / 2, cy / 2);
            ctx.rotate(degrees * Math.PI / 180 + Math.PI);
            ctx.beginPath();
            ctx.strokeStyle = color;
            ctx.lineWidth = lineWidth;
            ctx.moveTo(0, 0);
            ctx.lineTo(0, r - 2 * scale);
            ctx.moveTo(-1 * scale, r - 5 * scale);
            ctx.lineTo(0, r - 2 * scale);
            ctx.lineTo(1 * scale, r - 5 * scale);
            ctx.fillStyle = color;
            ctx.closePath();
            ctx.stroke();

            ctx.restore();
        },
    },
    watch: {
    }
});