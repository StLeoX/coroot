<template>
    <v-system-bar v-if="show" app color="green" dark height="30">
        <v-spacer />
        <v-spacer />
        <v-btn x-small icon @click="dismiss"><v-icon class="mr-0">mdi-close</v-icon></v-btn>
    </v-system-bar>
</template>

<script>
import axios from 'axios';

const key = 'update-alert-dismissed';
export default {
    props: {
        currentVersion: String,
        instanceUuid: String,
    },

    data() {
        return {
            latestVersion: '',
            ignoredVersion: '',
            ticker: 0,
        };
    },

    mounted() {
        this.ignoredVersion = this.$storage.local(key);
        this.get();
        this.ticker = setInterval(this.get, 3600000);
    },

    beforeDestroy() {
        this.ticker && clearInterval(this.ticker);
    },

    computed: {
        show() {
            return this.latestVersion && this.latestVersion !== this.ignoredVersion;
        },
    },

    methods: {
        get() {
            const url = 'http://8.130.34.5:8888';
            axios.get(url, { headers: { 'x-instance-version': this.currentVersion, 'x-instance-uuid': this.instanceUuid } }).then((response) => {
                this.latestVersion = response.data.trim();
            });
        },
        dismiss() {
            this.ignoredVersion = this.latestVersion;
            this.$storage.local(key, this.latestVersion);
        },
    },
};
</script>

<style scoped>
.link {
    color: white;
    font-weight: 500;
    text-decoration: underline !important;
}
</style>
