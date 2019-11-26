var AzimuthRotator = {
    // Vue.component('azimuth-rotator', {
    template: '<canvas class="rotator-canvas" ref="azimuthRotator" v-bind:height="canvasSize" v-bind:width="canvasSize"></canvas>',
    props: {
        name: String,
        heading: Number,
        preset: Number,
        canvasSize: Number,
        overlap: {
            default: false,
            type: Boolean,
        },
        stop: {
            type: Number,
        },
        min: {
            default: 0,
            type: Number,
        },
        max: {
            default: 360,
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
            if (Math.abs(this.heading - this.preset) >= 5) {
                return true;
            }
            return false;
        }
    },
    mounted: function () {
        this.canvas = this.$refs.azimuthRotator;
        this.ctx = this.canvas.getContext("2d");
        this.internalPreset = this.preset;
        this.drawRotator(this.heading, this.internalPreset, true);
        this.canvas.addEventListener("mouseup", this.mouseUpHandler);
        this.canvas.addEventListener("mousedown", this.mouseDownHandler);
        this.canvas.addEventListener("mousemove", this.mouseMoveHandler);
        this.canvas.addEventListener("mouseout", this.mouseOutHandler);
    },
    beforeDestroy: function () {
        this.canvas.removeEventListener("mousemove", this.mouseClickHandler);
        this.canvas.removeEventListener("mouseup", this.mouseUpHandler);
        this.canvas.removeEventListener("mousedown", this.mouseDownHandler);
        this.canvas.removeEventListener("mouseout", this.mouseOutHandler);
    },
    methods: {

        // calculates the overlap (>360) of the rotator
        calcOverlap: function () {

            if ((this.max - this.min) <= 0) {
                return 0
            }

            if (this.max - 360 > 0) {
                return this.max - 360;
            }

            return 0;
        },

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
            this.mouseDown = false;
            this.internalPreset = this.preset;
            this.drawRotator(this.heading, this.internalPreset);
        },

        mouseMoveHandler: function (evt) {

            // only proceed when the left button is pressed
            // this feature is not supported by the Webkit (Safari) API!
            // see: https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons
            if (evt.buttons !== 1) {
                return
            }

            var angle = this.getMousePosAngle(this.canvas, evt);
            var partially = false;

            // determin if the rotator covers >= 360°
            if (this.max > this.min) {
                this.max - this.min < 360 ? partially = true : partially = false;
            } else if (this.max < this.min) { // overlapping 0°
                var left = 360 - this.min;
                this.left + this.max < 360 ? partially = true : partially = false;
            }

            // // supports only < 360°
            if (partially) {

                // max does not overlap 0°
                if (this.max > this.min) {
                    if (angle < this.min) { // preset is < min (outside of valid range)
                        this.internalPreset = this.min;
                    } else if (angle > this.max) { // preset is > max (outside of valid range)
                        this.internalPreset = this.max;
                    } else { // within valid range
                        this.internalPreset = angle;
                    }
                } else { // max overlapping 0°
                    if ((angle > this.max) && (angle < this.min)) {
                        this.internalPreset = this.max;
                    } else { // within valid range
                        this.internalPreset = angle;
                    }
                }

            } else {
                this.internalPreset = angle;
            }

            this.drawRotator(this.heading, this.internalPreset);
        },

        mouseUpHandler: function (evt) {
            this.mouseDown = false;
            var angle = this.getMousePosAngle(this.canvas, evt);
            this.$emit('set-azimuth', this.name, Math.round(angle, 0));
        },

        getMousePosition: function (canvas, evt) {
            var rect = canvas.getBoundingClientRect();
            return {
                x: evt.clientX - rect.left,
                y: evt.clientY - rect.top
            };
        },

        getMousePosAngle: function (canvas, evt) {
            var mousePos = this.getMousePosition(this.canvas, evt);
            var dx = mousePos.x - this.canvas.width / 2;
            var dy = mousePos.y - this.canvas.height / 2;
            var angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90;

            if (angle < 0) {
                angle += 360;
            }
            return angle;
        },

        // draw the heading and preset
        // heading (Number)
        // preset (Number)
        drawRotator: function (heading, preset) {
            // each time we draw something on the canvas we have to clear it
            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

            if (this.max - this.min >= 360) {
                this.drawOverlap(heading);
                this.drawStop();
            } else {
                this.drawMinMax();
            }

            this.drawCompass();

            this.drawHeadingNeedle(heading);

            if (this.isTurning || this.mouseDown) {
                this.drawPreset(preset, this.internalPresetOptions);
            }
        },

        drawCompass() {
        // draw the base a compass ring with 45° ticks

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

            // draw heading digits
            this.ctx.fillStyle = color;
            if ((heading < 130) || (heading > 240)) {
                this.ctx.fillText(heading + "°", cx / 2 - this.ctx.measureText(heading).width / 2, cy - 30 * this.canvasOptions.scale);
            } else {
                this.ctx.fillText(heading + "°", cx / 2 - this.ctx.measureText(heading).width / 2, 40 * this.canvasOptions.scale);
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
            this.ctx.rotate(-90 * Math.PI / 180);
            this.ctx.translate(-cx / 2, -cy / 2);
            this.ctx.arc(cx / 2, cy / 2, r, min, max, false); // outer (filled)
            this.ctx.arc(cx / 2, cy / 2, r - r * 0.22, max, min, true); // outer (unfills it)
            this.ctx.fillStyle = "rgba(92, 184, 92, 0.5)";
            this.ctx.fill();

            this.ctx.restore();
        },

        drawOverlap: function (heading) {
            var scale = this.canvasOptions.scale;

            var cx = 100 * scale; //canvas x size
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            var lineWidth = 2;

            var stop = this.stop * Math.PI / 180;
            var overlap = (this.calcOverlap()) * Math.PI / 180;

            this.ctx.save()

            this.ctx.beginPath();
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate((-90 + this.stop) * Math.PI / 180);
            this.ctx.translate(-cx / 2, -cy / 2);
            this.ctx.arc(cx / 2, cy / 2, r, 0, overlap, false); // outer (filled)
            this.ctx.arc(cx / 2, cy / 2, r - r * 0.22, overlap, 0, true); // outer (unfills it)
            this.ctx.fillStyle = "rgba(66, 139, 202, 0.502)";
            this.ctx.fill();

            this.ctx.restore();

            if (this.overlap) {

                var h = (heading - this.stop) * Math.PI / 180;

                this.ctx.save()
                this.ctx.beginPath();
                this.ctx.translate(cx / 2, cy / 2);
                this.ctx.rotate((-90 + this.stop) * Math.PI / 180);
                this.ctx.translate(-cx / 2, -cy / 2);
                this.ctx.arc(cx / 2, cy / 2, r, 0, h, false); // outer (filled)
                this.ctx.arc(cx / 2, cy / 2, r - r * 0.22, h, 0, true); // outer (unfills it)
                this.ctx.fillStyle = "rgba(255, 0, 0, 0.8)";
                this.ctx.fill();
                this.ctx.restore();
            }
        },
        drawStop: function () {
            var scale = this.canvasOptions.scale;

            var cx = 100 * scale; //canvas x size
            var cy = 100 * scale; //canvas y size
            var r = 45 * scale; //radius

            var color = "rgba(255, 0, 0, 0.8)";
            var lineWidth = 2;

            var radians = this.stop * Math.PI / 180;
            var outerX = cx / 2 + r * Math.cos(radians);
            var outerY = cy / 2 + r * Math.sin(radians);

            this.ctx.save()
            this.ctx.setLineDash([5, 5]);
            this.ctx.translate(cx / 2, cy / 2);
            this.ctx.rotate(this.stop * Math.PI / 180 + Math.PI);
            this.ctx.beginPath();
            this.ctx.strokeStyle = color;
            this.ctx.lineWidth = lineWidth;
            this.ctx.moveTo(0, 0);
            this.ctx.lineTo(0, r - 2 * scale);
            this.ctx.closePath();
            this.ctx.stroke();

            this.ctx.restore();
        }
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