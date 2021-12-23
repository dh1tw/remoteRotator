var ElevationRotator = {
    // Vue.component('azimuth-rotator', {
    template: '<canvas class="rotator-canvas" ref="elevationRotator" v-bind:height="canvasSize" v-bind:width="canvasSize"></canvas>',
    props: {
        name: String,
        heading: Number,
        preset: Number,
        canvasSize: Number,
        min: {
            default: 0,
            type: Number,
        },
        max: {
            default: 180,
            type: Number,
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
            internalPreset: 0, // internal Preset
            mouseDown: false,
            touchOngoing: false,
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
        this.canvas = this.$refs.elevationRotator;
        this.ctx = this.canvas.getContext("2d");
        this.internalPreset = this.preset;
        this.drawRotator(this.heading, this.internalPreset, true);
        this.canvas.addEventListener("mouseup", this.mouseUpHandler);
        this.canvas.addEventListener("mousedown", this.mouseDownHandler);
        this.canvas.addEventListener("mousemove", this.mouseMoveHandler);
        this.canvas.addEventListener("mouseout", this.mouseOutHandler);
        this.canvas.addEventListener("touchstart", this.touchStartHandler, false);
        this.canvas.addEventListener("touchmove", this.touchMoveHandler, false);
        this.canvas.addEventListener("touchend", this.touchEndHandler, false);
    },
    beforeDestroy: function () {
        this.canvas.removeEventListener("mousemove", this.mouseClickHandler);
        this.canvas.removeEventListener("mouseup", this.mouseUpHandler);
        this.canvas.removeEventListener("mousedown", this.mouseDownHandler);
        this.canvas.removeEventListener("mouseout", this.mouseOutHandler);
        this.canvas.removeEventListener("touchstart", this.touchStartHandler);
        this.canvas.removeEventListener("touchmove", this.touchMoveHandler);
        this.canvas.removeEventListener("touchend", this.touchEndHandler);
    },
    methods: {
        // calculate the width of the heading needle (depends on the canvas size)
        headingNeedleWidth: function () {
            if (this.canvasSize > 100) {
                return this.canvasSize / 30;
            }
            return 7
        },
        // calculate the width of the preset needle (depends on the canvas size)
        presetNeedleWidth: function () {
            if (this.canvasSize > 100) {
                return this.canvasSize / 80;
            }
            return 3
        },
        // calculate the font size (depends on the canvas size)
        headingFont: function () {
            return "normal " + this.canvasSize / 15 + "pt Inconsolata";
        },
        mouseDownHandler: function (evt) {
            this.mouseDown = true;
        },
        mouseOutHandler: function (evt) {
            // if the cursor moves out of the canvas, then
            // ignore the preset
            this.mouseDown = false;
            this.internalPreset = this.preset;
            this.drawRotator(this.heading, this.internalPreset);
        },
        mouseMoveHandler: function (evt) {
            // only proceed when the left button is pressed
            if (evt.buttons !== 1) {
                return
            }
            this.calculatePreset(evt);
        },
        mouseUpHandler: function (evt) {
            this.mouseDown = false;
            this.calculatePreset(evt);
            this.$emit('set-elevation', this.name, Math.round(this.internalPreset, 0));
        },
        touchStartHandler: function (evt) {
            this.touchOngoing = true;
        },
        touchMoveHandler: function (evt) {
            this.calculatePreset(evt);
        },
        touchEndHandler: function (evt) {
            this.touchOngoing = false;
            this.$emit('set-elevation', this.name, Math.round(this.internalPreset, 0));
        },
        calculatePreset: function (evt) {
            var p = this.getCursorPosAngle(this.canvas, evt);

            // only values between min and max are allowed
            if (p > this.max && p <= 270) {
                p = this.max;
            } else if (p > 270) {
                p = this.min;
            } else if (p < this.min) {
                p = this.min;
            }

            // no need to redraw
            if (p === this.internalPreset) {
                return
            }

            this.internalPreset = p;

            this.drawRotator(this.heading, this.internalPreset);
        },
        getCursorPosition: function (canvas, evt) {
            var rect = canvas.getBoundingClientRect();
            if ("touches" in evt) { // only touch events have the property 'touches'
                return {
                    x: evt.touches[0].clientX - rect.left,
                    y: evt.touches[0].clientY - rect.top
                }
            }

            return { // must be a mouse event 
                x: evt.clientX - rect.left,
                y: evt.clientY - rect.top
            }
        },
        getCursorPosAngle: function (canvas, evt) {
            var cursorPos = this.getCursorPosition(this.canvas, evt);
            var dx = cursorPos.x - this.canvas.width / 2;
            var dy = cursorPos.y - this.canvas.height / 2;
            var angle = Math.atan2(dy, dx) * (180 / Math.PI) + 180;

            return angle;
        },
        // draw the heading and preset.
        // heading (Number)
        // preset (Number)
        drawRotator: function (heading, preset) {
            // each time we draw something on the canvas we have to clear it
            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

            if (Math.abs(Math.abs(this.max) - Math.abs(this.min)) < 180) {
                this.drawMinMax();
            }

            this.drawCompass();

            this.drawHeadingNeedle(heading);

            if ((Math.round(preset) == heading) && !this.isTurning) {
                return
            }

            if (this.isTurning || this.mouseDown || this.touchOngoing) {
                this.drawPreset(preset, this.internalPresetOptions);
            }

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
            this.ctx.save();

            this.ctx.beginPath();
            this.ctx.strokeStyle = color;
            this.ctx.lineWidth = lineWidth;
            this.ctx.arc(cx / 2, cy / 2, r, 0, Math.PI, true);
            this.ctx.stroke();
            this.ctx.closePath();

            //draw 45° and 15° ticks
            this.ctx.translate(cx / 2, cy / 2);
            for (i = -90; i <= 90; i++) {
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
                var h = "M"; // since we use Monospace font, all letters have the same vertical/horizontal size
                var txt = "90°";
                this.ctx.fillText(txt, cx / 2 - this.ctx.measureText(txt).width / 2, 25 * this.canvasOptions.scale);
                txt = "0°"
                this.ctx.fillText(txt, 18 * this.canvasOptions.scale, (cy / 2) + this.ctx.measureText(h).width / 2);
                txt = "180°"
                this.ctx.fillText(txt, cx - 33 * this.canvasOptions.scale, (cy / 2) + this.ctx.measureText(h).width / 2);
            }
        },
        drawHeadingNeedle: function (heading) {

            var scale = this.canvasOptions.scale;

            var color = this.headingNeedleOptions.needleColor;
            var cx = 100 * scale; //canvas x size 
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            // draw heading digits
            this.ctx.fillStyle = color;
            this.ctx.fillText(heading + "°", cx / 2 - this.ctx.measureText(heading).width / 2, cy - 35 * this.canvasOptions.scale);

            this.ctx.save();

            var lineWidth = this.headingNeedleWidth();
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate(heading * Math.PI / 180 + Math.PI / 2);
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
            this.ctx.arc(cx / 2, cy / 2, lineWidth / 2, 0, 2 * Math.PI);
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
            this.ctx.rotate(degrees * Math.PI / 180 + Math.PI / 2);
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
        drawMinMax: function (heading) {
            var scale = this.canvasOptions.scale;

            var cx = 100 * scale; //canvas x size 
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            var lineWidth = 2;

            var min = this.min * Math.PI / 180;
            var max = this.max * Math.PI / 180;

            this.ctx.save()

            this.ctx.beginPath();
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate(180 * Math.PI / 180);
            this.ctx.translate(-cx / 2, -cy / 2);
            this.ctx.arc(cx / 2, cy / 2, r, min, max, false); // outer (filled)
            this.ctx.arc(cx / 2, cy / 2, r - r * 0.22, max, min, true); // outer (unfills it)
            this.ctx.fillStyle = "rgba(92, 184, 92, 0.5)";
            this.ctx.fill();

            this.ctx.restore();
        },
    },
    watch: {
        heading: function () {
            this.drawRotator(this.heading, this.internalPreset);
        },
        preset: function () {
            if (!this.mouseDown) {
                this.internalPreset = this.preset;
                this.drawRotator(this.heading, this.internalPreset);
            }
        },
        min: function () {
            this.drawRotator(this.heading, this.internalPreset);
        },
        max: function () {
            this.drawRotator(this.heading, this.internalPreset);
        },
        stop: function () {
            this.drawRotator(this.heading, this.internalPreset);
        },
        canvasSize: function (newCanvasSize) {
            this.$set(this.canvasOptions, "scale", newCanvasSize / 100);

            // wait one tick until the canvas has been re-initialized
            Vue.nextTick(() => {
                this.drawRotator(this.heading, this.internalPreset);
            });
        }
    }
}