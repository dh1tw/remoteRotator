var vm = new Vue({
    el: '#app',

    data: {
        ws: null, // websocket
        rotators: {},
        selectedAzRotator: {
            name: "n/a",
            azimuth: 0,
            az_preset: 0,
            az_stop: 0,
            az_min: 0,
            az_max: 360,
            az_overlap: false,
        },
        selectedElRotator: {
            name: "n/a",
            elevation: 0,
            el_min: 0,
            el_max: 180,
            el_preset: 0,
        },
        canvasSize: 200,
        hideConnectionMsg: false,
        resizeTimeout: null,
        connected: false,
    },
    components: {
        'azimuth-rotator': AzimuthRotator,
        'elevation-rotator': ElevationRotator,
        'rotator-name': RotatorName,
    },
    created: function () {
        window.addEventListener('resize', this.getWindowSize);
        this.resizeWindow();
    },
    mounted: function () {
        this.openWebsocket();
    },
    beforeDestroy: function () {
        window.removeEventListener('resize', this.getWindowWidth);
    },
    methods: {

        // add a rotator
        addRotator: function (rotator) {

            if (!(rotator.name in this.rotators)) {
                this.$set(this.rotators, rotator.name, rotator);
            }

            // if this is the first rotator, set the main azimuth rotator component
            if (this.selectedAzRotator.name === "n/a" && rotator.has_azimuth) {
                this.selectedAzRotator = rotator;
            }
            if (this.selectedElRotator.name === "n/a" && rotator.has_elevation) {
                this.selectedElRotator = rotator;
            }
            this.resizeWindow();
        },

        // remove a rotator
        removeRotator: function (rotator) {

            if (rotator.name in this.rotators) {

                this.$delete(this.rotators, rotator.name);

                // check if other azimuth rotators are still available
                if (Object.keys(this.azRotators).length > 0) {
                    if (this.selectedAzRotator.name == rotator.name) {
                        // pick the first one in the list
                        var nextRot = Object.keys(this.azRotators)[0];
                        this.selectedAzRotator = this.azRotators[nextRot];
                    }
                } else {
                    // no more rotators left
                    this.selectedAzRotator = {
                        name: "n/a",
                        azimuth: 0,
                        az_min: 0,
                        az_max: 360,
                        az_overlap: false,
                        az_preset: 0,
                        az_stop: 0,
                    }
                }

                // check if other elevation rotators are still available
                if (Object.keys(this.elRotators).length > 0) {
                    if (this.selectedElRotator.name == rotator.name) {
                        // pick the first one in the list
                        var nextRot = Object.keys(this.azRotators)[0];
                        this.selectedElRotator = this.elRotators[nextRot];
                    }
                } else {
                    // no more rotators left
                    this.selectedElRotator = {
                        name: "n/a",
                        elevation: 0,
                        el_min: 0,
                        el_max: 180,
                        el_preset: 0,                    }
                }
            }

            this.resizeWindow();
        },

        // open the websocket and set an eventlister to receive updates
        // for rotators
        openWebsocket: function () {
            this.ws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');
            this.ws.addEventListener('message', function (e) {
                var eventMsg = JSON.parse(e.data);
                // console.log(eventMsg);

                if (eventMsg['name'] == 'add') {
                    this.addRotator(eventMsg['rotator']);
                } else if (eventMsg['name'] == 'remove') {
                    this.removeRotator(eventMsg['rotator']);
                } else if (eventMsg['name'] == 'heading') {
                    newHeading = eventMsg['status']
                    if (newHeading.name in this.rotators) {
                        // copy values
                        var rotator = this.rotators[newHeading.name]
                        if (rotator.has_azimuth) {
                            this.$set(this.rotators[newHeading.name], 'azimuth', newHeading.azimuth);
                            this.$set(this.rotators[newHeading.name], 'az_preset', newHeading.az_preset);
                            this.$set(this.rotators[newHeading.name], 'az_overlap', newHeading.az_overlap);
                        }
                        if (rotator.has_elevation) {
                            this.$set(this.rotators[newHeading.name], 'elevation', newHeading.elevation);
                            this.$set(this.rotators[newHeading.name], 'el_preset', newHeading.el_preset);
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
                for (rotator in this.rotators){
                    this.removeRotator(this.rotators[rotator]);
                }
                this.rotators = {}
            }.bind(this));
        },

        // set the active azimuth rotator
        setAzRotator: function (name) {
            if (name in this.rotators) {
                this.selectedAzRotator = this.rotators[name]
            }
        },

        // set the active elevation rotator
        setElRotator: function (name) {
            if (name in this.rotators) {
                this.selectedElRotator = this.rotators[name]
            }
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
        // send a request to the server to set elevation
        setElevation: function (name, heading) {
            var msg = {
                "name": name,
                "has_elevation": true,
                "elevation": heading,
            }
            var data = JSON.stringify(msg);
            this.ws.send(data);
        },

        // helper funtion for resizing window. This function reduces
        // the amount of resize events to just one.
        getWindowSize: function () {
            clearTimeout(this.resizeTimeout);
            this.resizeTimeout = setTimeout(this.resizeWindow, 400);
        },

        resizeWindow: function (event) {

            var width = document.documentElement.clientWidth;
            var height = document.documentElement.clientHeight;

            // azimuth and elevation rotators available
            console.log("width:" + width);
            console.log("height:" + height);

            // azimuth AND elevation rotator available
            if (Object.keys(this.azRotators).length > 0 && Object.keys(this.elRotators).length > 0) {
                if (width > height) {
                    this.canvasSize = width * 2 / 5;
                    if (this.canvasSize > height) {
                        this.canvasSize = height * 4 / 5;
                    }
                } else {
                    this.canvasSize = width / 2 - width/10;
                    if (this.canvasSize < 200){
                        this.canvasSize = 300;
                    }
                }
                // only azimuth or elevation rotator available
            } else {
                if (width > height) {
                    this.canvasSize = height - 120;
                } else {
                    this.canvasSize = width - 70;
                }
            }
            console.log("canvas:" + this.canvasSize);
        },
    },
    computed: {
        // returns an object containing all azimuth rotators
        azRotators: function ()  {
            var rotators = this.rotators;
            var azRotators = {};
            Object.keys(rotators).forEach(function (key) {
                if (rotators[key].has_azimuth) {
                    azRotators[key] = rotators[key];
                }
            });
            return azRotators;
        },
        // returns an object containing all elevation rotators
        elRotators: function ()  {
            var rotators = this.rotators;
            var elRotators = {};
            Object.keys(rotators).forEach(function (key) {
                if (rotators[key].has_elevation) {
                    elRotators[key] = rotators[key];
                }
            });
            return elRotators;
        },

        // returns all azimuth rotators, ordered alphabetically
        sortedAzRotators: function () {
            var azRotators = this.azRotators;
            const ordered = {};
            Object.keys(azRotators).sort().forEach(function (key) {
                ordered[key] = azRotators[key];
            });
            return ordered;
        },
        // returns all elevation rotators, ordered alphabetically
        sortedElRotators: function () {

            var elRotators = this.elRotators;
            const ordered = {};
            Object.keys(elRotators).sort().forEach(function (key) {
                ordered[key] = elRotators[key];
            });
            return ordered;
        },
    }
});