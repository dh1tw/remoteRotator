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
            az_max: 450,
            elevation: 0,
            el_preset: 0,
            has_azimuth: true,
            has_elevation: false,
        },
        azEnabled: false,
        azCanvasSize: 400,
        hideConnectionMsg: false,
        resizeTimeout: null,
    },
    components: {
        'azimuth-rotator': AzimuthRotator,
    },
    created: function () {
        window.addEventListener('resize', this.getWindowSize);
        this.resizeWindow();
    },
    mounted: function () {
        this.openWebsocket();
    },
    methods: {

        // add a rotator
        addRotator: function (rotator) {

            if (!(rotator.name in this.rotators)) {
                this.$set(this.rotators, rotator.name, rotator);
                console.log(rotator);
            }

            // if this is the first rotator, set the main azimuth rotator component
            if (this.selectedAzRotator.name === "n/a" && rotator.has_azimuth) {
                this.azEnabled = true;
                this.selectedAzRotator = rotator;
            }
        },

        // remove a rotator
        removeRotator: function (rotator) {

            if (rotator.name in this.rotators) {
                
                this.$delete(this.rotators, rotator.name);

                if (Object.keys(this.rotators).length > 0){
                    if (this.selectedAzRotator.name == rotator.name){
                        this.selectedAzRotator = this.rotators[Object.keys(this.rotators)[0]];
                    }                
                } else {
                    // no more rotators left
                    this.selectedAzRotator = {
                        name: "n/a",
                        azimuth: 0,
                        az_preset: 0,
                        az_stop: 0,
                        elevation: 0,
                        el_preset: 0,
                        has_azimuth: true,
                        has_elevation: false,
                    }
                    this.azEnabled = false;
                }
            }
        },

        // open the websocket and set an eventlister to receive updates
        // for rotators
        openWebsocket: function () {
            this.ws = new ReconnectingWebSocket('ws://' + window.location.host + '/ws');
            this.ws.addEventListener('message', function (e) {
                var eventMsg = JSON.parse(e.data);

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
            }.bind(this));
        },

        // set the active azimuth rotator
        setAzRotator: function (name) {
            if (name in this.rotators) {
                this.selectedAzRotator = this.rotators[name]
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

        // helper funtion for resizing window. This function reduces
        // the amount of resize events to just one.
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
            this.$forceUpdate();
        },
    },
    beforeDestroy() {
        window.removeEventListener('resize', this.getWindowWidth);
    },
    computed: {
        // order the rotators alphabetically
        sortedRotators: function () {
            var rotators = this.rotators;

            const ordered = {};
            Object.keys(rotators).sort().forEach(function (key) {
                ordered[key] = rotators[key];
            });
            return ordered;
        },
    }
});