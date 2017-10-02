var AzimuthRotator = {
    // Vue.component('azimuth-rotator', {
    template: '<canvas class="rotator-canvas" ref="azimuthRotator" v-bind:height="canvasSize" v-bind:width="canvasSize"></canvas>',
    props: {
        name: String,
        heading: Number,
        preset: Number,
        enabled: Boolean,
        canvasSize: Number,
        min: Number,
        max: Number,
        // stop: {
        //     default: 225,
        //     type: Number,
        // },
        showHeading: {
            default: true,
            type: Boolean,
        },
        showLegend: {
            default: true,
            type: Boolean,
        },
    },
    data: function () {
        return {
            canvas: null,
            ctx: null,
            canvasOptions: {
                scale: this.canvasSize / 100,
                color: "#FFF",
                lineWidth: 3,
            },
            headingNeedleOptions: {
                needleColor: "red",
                needleColorRing: "yellow",
            },
            presetNeedleOptions: {
                color: "yellow",
            }
        }
    },
    computed: {
        // returns if the rotator is turning
        isTurning: function () {
            if (Math.abs(this.heading - this.preset) >= 3) {
                return true;
            }
            return false;
        }
    },
    mounted: function () {
        this.canvas = this.$refs.azimuthRotator;
        this.ctx = this.canvas.getContext("2d");
        this.drawRotator(this.heading, this.preset, true);
        this.addCanvasClickHandler();
    },
    methods: {

        // calculate the width of the heading needle (depends on the canvas size)
        headingNeedleWidth: function(){
            if (this.canvasSize > 100) {
                return this.canvasSize / 30;
            }
            return 7
        },

        // calculate the width of the preset needle (depends on the canvas size)
        presetNeedleWidth: function(){
            if (this.canvasSize > 100) {
                return this.canvasSize / 80;
            }
            return 3
        },

        // calculate the font size (depends on the canvas size)
        headingFont: function(){
            return "normal " + this.canvasSize / 15 + "pt Inconsolata";
        },

        // draw the heading and preset. Through needlesEnabled the heading and
        // preset needle are enabled/disabled.
        // heading (Number)
        // preset (Number)
        // needlesEnabled (boolean)
        drawRotator: function (heading, preset, needlesEnabled) {
            // each time we draw something on the canvas we have to clear it
            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

            this.drawCompass();
            
            // if (needlesEnabled) {
            //     this.drawStop();
            // }

            if (needlesEnabled) {
                this.drawHeadingNeedle(heading);
            }

            if (this.isTurning && needlesEnabled) {
                this.drawPreset(preset, this.presetOptions);
            }
        },

        // add a listener for clicks on the canvas. The listener will 
        // retrieve the selected heading and emit an event with 
        // name and heading as arguments.
        addCanvasClickHandler: function () {

            this.canvas.addEventListener('click', function (evt) {
                var mousePos = this.getMousePosition(this.canvas, evt);
                var dx = mousePos.x - this.canvas.width / 2;
                var dy = mousePos.y - this.canvas.height / 2;
                var angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90;

                if (angle < 0) {
                    angle += 360;
                }

                // console.log(angle);
                this.$emit('set-azimuth', this.name, Math.round(angle, 0));

            }.bind(this), false);
        },

        getMousePosition: function (canvas, evt) {
            var rect = canvas.getBoundingClientRect();
            return {
                x: evt.clientX - rect.left,
                y: evt.clientY - rect.top
            };
        },

        // draw the base a compass ring with 45° ticks
        drawCompass() {

            var cx = 100 * this.canvasOptions.scale; //canvas x size 
            var cy = 100 * this.canvasOptions.scale; //canvas y size
            var r = 45 * this.canvasOptions.scale; //radius
            var font = this.headingFont();
            var color = this.canvasOptions.color;
            var needleColor = this.canvasOptions.needleColor;
            var needleColorRing = this.canvasOptions.needleColorRing;
            var lineWidth = this.canvasOptions.lineWidth;

            var lstx = r - (r * 0.22); //large tick start x
            var lsty = -r + (r * 0.22); //large tick start y
            var letx = r; //large tick end x
            var lety = -r; //large tick end y    
            var sstx = r - (r * 0.11); //small tick start x
            var ssty = -r + (r * 0.11); //small tick start y
            var setx = r; //small tick end x
            var sety = -r; //small tick end y

            //outer ring
            this.ctx.beginPath();
            this.ctx.strokeStyle = color;
            this.ctx.lineWidth = lineWidth;
            this.ctx.arc(cx / 2, cy / 2, r, 0, 2 * Math.PI);
            this.ctx.stroke();
            this.ctx.closePath();

            this.ctx.save();

            //draw 45° and 15° ticks
            this.ctx.translate(cx / 2, cy / 2);
            for (i = 1; i <= 360; i++) {
                ang = Math.PI / 180 * i;
                sang = Math.sin(ang);
                cang = Math.cos(ang);
                //If modulus of divide by 45 is zero then draw an degree marker + numeral
                if (i % 45 == 0) {
                    this.ctx.lineWidth = 3;
                    sx = sang * lstx;
                    sy = cang * lsty;
                    ex = sang * letx;
                    ey = cang * lety;
                    this.ctx.beginPath();
                    this.ctx.moveTo(sx, sy);
                    this.ctx.lineTo(ex, ey);
                    this.ctx.stroke();
                }
                //Else draw every 10deg a small degree marker
                else if (i % 15 == 0) {
                    this.ctx.lineWidth = 1;
                    sx = sang * sstx;
                    sy = cang * ssty;
                    ex = sang * setx;
                    ey = cang * sety;
                    this.ctx.beginPath();
                    this.ctx.moveTo(sx, sy);
                    this.ctx.lineTo(ex, ey);
                    this.ctx.stroke();
                }
            }

            this.ctx.restore();

            //North East South West Labels
            if (this.showLegend) {
                this.ctx.font = font;
                this.ctx.fillStyle = color;
                var txt = "M"; // since we use Monospace font, all letters have the same vertical/horizontal size
                this.ctx.fillText("N", cx / 2 - this.ctx.measureText(txt).width / 2, 25 * this.canvasOptions.scale);
                this.ctx.fillText("S", cx / 2 - this.ctx.measureText(txt).width / 2, cy - 17 * this.canvasOptions.scale);
                this.ctx.fillText("W", 16 * this.canvasOptions.scale, (cy / 2) + this.ctx.measureText(txt).width / 2);
                this.ctx.fillText("E", cx - 22 * this.canvasOptions.scale, (cy / 2) + this.ctx.measureText(txt).width / 2);
            }
        },

        drawHeadingNeedle: function (heading) {

            var scale = this.canvasOptions.scale;
            
            var color = this.headingNeedleOptions.needleColor;
            var cx = 100 * scale; //canvas x size 
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            // draw heading digits if enabled
            if (this.showHeading) {
                this.ctx.fillStyle = color;
                if ((heading < 130) || (heading > 240)) {
                    this.ctx.fillText(heading + "°", cx / 2 - this.ctx.measureText(heading).width / 2, cy - 30 * this.canvasOptions.scale);
                } else {
                    this.ctx.fillText(heading + "°", cx / 2 - this.ctx.measureText(heading).width / 2, 40 * this.canvasOptions.scale);
                }
            }

            this.ctx.save();
            var lineWidth = this.headingNeedleWidth(); 
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate(heading * Math.PI / 180 + Math.PI);
            this.ctx.beginPath();
            this.ctx.moveTo(-lineWidth, 0);
            this.ctx.lineTo(0, r);
            this.ctx.lineTo(lineWidth, 0);

            this.ctx.fillStyle = this.headingNeedleOptions.needleColor;
            this.ctx.closePath();
            this.ctx.fill();

            this.ctx.restore();

            //outer ring around compass needle
            this.ctx.beginPath();
            this.ctx.arc(cx / 2, cy / 2, lineWidth, 0, 2 * Math.PI);
            this.ctx.fillStyle = this.headingNeedleOptions.needleColor;
            this.ctx.fill();
            this.ctx.closePath();

            if (this.canvasOptions.drawDigits != null) {
                var drawDigits = this.canvasOptions.drawDigits;
            } else {
                var drawDigits = true;
            }

            //inner ring around compass needle
            this.ctx.translate(0, 0);
            this.ctx.beginPath();
            this.ctx.arc(cx / 2, cy / 2, lineWidth/2, 0, 2 * Math.PI);
            this.ctx.fillStyle = this.headingNeedleOptions.needleColorRing;
            this.ctx.fill();
            this.ctx.closePath();
        },

        drawPreset: function (degrees) {
            var scale = this.canvasOptions.scale;

            var cx = 100 * scale; //canvas x size 
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            var font = this.presetNeedleOptions.font;
            var color = this.presetNeedleOptions.color;
            var lineWidth = this.presetNeedleWidth();

            var radians = degrees * Math.PI / 180;
            var outerX = cx / 2 + r * Math.cos(radians);
            var outerY = cy / 2 + r * Math.sin(radians);

            this.ctx.save()
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate(degrees * Math.PI / 180 + Math.PI);
            this.ctx.beginPath();
            this.ctx.strokeStyle = color;
            this.ctx.lineWidth = lineWidth;
            this.ctx.moveTo(0, 0);
            this.ctx.lineTo(0, r - 2 * scale);
            this.ctx.moveTo(-1 * scale, r - 5 * scale);
            this.ctx.lineTo(0, r - 2 * scale);
            this.ctx.lineTo(1 * scale, r - 5 * scale);
            this.ctx.fillStyle = color;
            this.ctx.closePath();
            this.ctx.stroke();

            this.ctx.restore();
        },

        // drawStop: function () {
        //     var scale = this.canvasOptions.scale;

        //     var cx = 100 * scale; //canvas x size 
        //     var cy = 100 * scale; //canvas y size
        //     var r = 45 * scale; //radius

        //     var color = "blue";
        //     var lineWidth = 2;

        //     var radians = this.stop * Math.PI / 180;
        //     var outerX = cx / 2 + r * Math.cos(radians);
        //     var outerY = cy / 2 + r * Math.sin(radians);

        //     this.ctx.save()
        //     this.ctx.setLineDash([5, 5]);
        //     this.ctx.translate(cx / 2, cy / 2);
        //     this.ctx.rotate(this.stop * Math.PI / 180 + Math.PI);
        //     this.ctx.beginPath();
        //     this.ctx.strokeStyle = color;
        //     this.ctx.lineWidth = lineWidth;
        //     this.ctx.moveTo(0, 0);
        //     this.ctx.lineTo(0, r - 2 * scale);
        //     // this.ctx.moveTo(-1 * scale, r - 5 * scale);
        //     // this.ctx.lineTo(0, r - 2 * scale);
        //     // this.ctx.lineTo(1 * scale, r - 5 * scale);
        //     // this.ctx.fillStyle = color;
        //     this.ctx.closePath();
        //     this.ctx.stroke();

        //     this.ctx.restore();
        // }
    },
    watch: {
        heading: function (newHeading, oldHeading) {
            this.drawRotator(this.heading, this.preset, this.enabled);
        },
        preset: function (newPreset) {
            this.drawRotator(this.heading, this.preset, this.enabled);
        },
        enabled: function (newEnabled) {
            this.drawRotator(this.heading, this.preset, this.enabled);
        },
        canvasSize: function (newCanvasSize) {
            console.log("new canvas size:", newCanvasSize);
            this.$set(this.canvasOptions, "scale", newCanvasSize / 100);

            // wait one tick until the canvas has been re-initialized
            Vue.nextTick(()=> {
                this.drawRotator(this.heading, this.preset, this.enabled);
            });
        }
    }
}