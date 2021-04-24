var RotatorName = {
    template: '<div class="rotator-name" :style="styleObj"><div class="tag">{{typeLabel}}</div><div class="name">{{name}}</div></div>',
    props: {
        name: String,
        isAzimuth: Boolean,
        width: Number,
    },
    mounted: function () {},
    beforeDestroy: function () {},
    methods: {},
    computed: {
        typeLabel: function () {
            if (this.isAzimuth) {
                return "AZ";
            }
            return "EL";
        },
        styleObj: function() {
            return {
                "max-width": this.width + 'px',
                "font-size": this.width / 15 + "pt",
            }
        }
    },
    watch: {},
}