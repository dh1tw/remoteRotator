Vue.component('azimuth-rotator', {
    template: '<div><canvas height=200 width=200></canvas></div>',
    props: {
        heading: Number,
        preset: Number,
        enabled: Boolean,
    },
    data : function(){
        return {
            counter : 0,
            enabled: false,
            heading: 0,
            preset: 0,
        }
    },
    mounted: function(){

    },
    methods: {
        incrementCounter: function(){
            this.counter += 1;
            this.$emit('increment')
        }
    }
})

// Vue.component('button-counter', {
//     template: '<button v-on:click="incrementCounter">{{ counter }}</button>',
//     data: function () {
//         return {
//             counter: 0
//         }
//     },
//     methods: {
//         incrementCounter: function () {
//             this.counter += 1
//             this.$emit('increment')
//         }
//     },
// })