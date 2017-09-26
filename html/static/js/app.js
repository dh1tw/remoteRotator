var vm = new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        rotators: {},
        azName: "n/a",
        azHeading: 0,
        azPreset: 0,
        azEnabled: false,
        azCanvasSize: 400,
        hideConnectionMsg: false,
        resizeTimeout: null,
    },
    components: {
        'azimuth-rotator': AzimuthRotator,
    },
    beforeCreate: function () {},
    created: function () {
        window.addEventListener('resize', this.getWindowSize);
        this.resizeWindow();
    },
    mounted: function () {
        this.getRotators();
    },
    methods: {

        // Make and Ajax request to the server to get a list
        // of all available rotators
        getRotators: function () {
            this.$http.get("/info").then(rotators => {
                rotatorInfo = JSON.parse(rotators.bodyText);
                if (Object.prototype.toString.call(rotatorInfo) === '[object Array]') {
                    for (i = 0; i < rotatorInfo.length; i++) {
                        this.addRotator(rotatorInfo[i]);
                    }
                } else {
                    this.addRotator(rotatorInfo);
                }

                // TBD check if a rotator has disappeared

                if (this.ws == null) {
                    this.openWebsocket();
                }
            });
        },

        // add a rotator
        addRotator: function (rotator) {

            this.rotators[rotator.name] = rotator;

            // if this is the first rotator, set the main azimuth rotator component
            if (this.azName === "n/a" && rotator.has_azimuth) {
                this.azName = rotator.name;
                this.azHeading = 0;
                this.azPreset = 0;
                this.azEnabled = false;
            }
        },

        // open the websocket and set an eventlister to receive updates
        // for rotators
        openWebsocket: function () {
            this.ws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');

            this.ws.addEventListener('message', function (e) {
                var rotatorsMsg = JSON.parse(e.data);

                for (i = 0; i < rotatorsMsg.length; i++) {
                    newRotator = rotatorsMsg[i]
                    if (newRotator.name in this.rotators) {
                        // copy values
                        var rotator = this.rotators[newRotator.name]
                        if (rotator.has_azimuth) {
                            rotator.azimuth = newRotator.azimuth;
                            rotator.az_preset = newRotator.az_preset;
                        }
                        if (rotator.has_elevation) {
                            rotator.elevation = newRotator.elevation;
                            rotator.el_preset = newRotator.el_preset;
                        }

                        // update the main azimuth rotator component
                        if (newRotator.name == this.azName) {
                            this.azHeading = newRotator.azimuth;
                            this.azPreset = newRotator.az_preset;
                            this.azEnabled = true;
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

        // send a request to the server to set azimuth
        setAzimuth: function (name, heading) {
            var msg = {
                "name": name,
                "has_azimuth": true,
                "azimuth": heading,
            }
            var data = JSON.stringify(msg);
            this.ws.send(data);
        },
        getWindowSize: function () {
            clearTimeout(this.resizeTimeout);
            this.resizeTimeout = setTimeout(this.resizeWindow, 400);
        },
        resizeWindow: function (event) {

            var width = document.documentElement.clientWidth;
            var height = document.documentElement.clientHeight;

            if (height < width) {
                this.azCanvasSize = document.documentElement.clientHeight - 120;
            } else {
                this.azCanvasSize = document.documentElement.clientWidth - 70;
            }
        }

    },
    beforeDestroy() {
        window.removeEventListener('resize', this.getWindowWidth);
    },
    watch: {

    }
});